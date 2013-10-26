package storage

import (
	"fmt"
	"github.com/syndtr/goleveldb/leveldb"
	"github.com/syndtr/goleveldb/leveldb/opt"
	"strings"
	"sync"
)

// 使用LevelDB做数据源
type LevelDBDataSource struct {
	GoRedisDataSource
	db    *leveldb.DB
	ro    *opt.ReadOptions
	wo    *opt.WriteOptions
	mutex sync.Mutex
}

func NewLevelDBDataSource(path string) (l *LevelDBDataSource, err error) {
	l = &LevelDBDataSource{}
	l.ro = &opt.ReadOptions{}
	l.wo = &opt.WriteOptions{}
	options := opt.Options{}
	options.SetFlag(opt.OFCreateIfMissing)
	options.SetMaxOpenFiles(20000)
	options.SetWriteBuffer(256 << 20)
	l.db, err = leveldb.OpenFile(path, &options)
	return
}

func (l *LevelDBDataSource) DB() *leveldb.DB {
	return l.db
}

func (l *LevelDBDataSource) Get(key []byte) (entry Entry) {
	// l.mutex.Lock()
	// defer l.mutex.Unlock()
	bs, e1 := l.db.Get(key, l.ro)
	if e1 != nil || len(bs) == 0 {
		return
	}

	switch EntryType(bs[0]) {
	case EntryTypeString:
		entry = NewStringEntry(nil)
	case EntryTypeHash:
		entry = NewHashEntry()
	case EntryTypeSortedSet:
		entry = NewSortedSetEntry()
	case EntryTypeSet:
		entry = NewSetEntry()
	case EntryTypeList:
		entry = NewListEntry()
	default:
		entry = NewStringEntry(nil)
	}
	// 反序列化
	err := entry.Decode(bs[1:])
	if err != nil {
		fmt.Println(err)
	}
	return
}

/*
	batch := new(leveldb.Batch)
	count := len(keyvals)
	for i := 0; i < count; i += 2 {
		batch.Put([]byte(keyvals[i]), []byte(keyvals[i+1]))
	}
	err = s.db.Write(batch, s.wo)
*/
func (l *LevelDBDataSource) Set(key []byte, entry Entry) (err error) {
	// l.mutex.Lock()
	// defer l.mutex.Unlock()
	var bs []byte
	bs, err = entry.Encode()
	if err == nil {
		buf := make([]byte, len(bs)+1)
		copy(buf, []byte{byte(entry.Type())})
		copy(buf[1:], bs)
		err = l.db.Put(key, buf, l.wo)
	}
	return
}

func (l *LevelDBDataSource) Keys(pattern string) (keys []string) {
	l.mutex.Lock()
	defer l.mutex.Unlock()
	keys = make([]string, 0, 100)
	iter := l.db.NewIterator(l.ro)
	if pattern != "*" {
		iter.Seek([]byte(pattern))
	}
	for iter.Next() {
		key := string(iter.Key())
		if pattern == "*" || strings.HasPrefix(key, pattern) {
			keys = append(keys, key)
		} else {
			break
		}
	}
	iter.Release()
	return
}

func (l *LevelDBDataSource) Remove(key []byte) (err error) {
	// l.mutex.Lock()
	// defer l.mutex.Unlock()
	err = l.db.Delete(key, l.wo)
	return
}

func (l *LevelDBDataSource) NotifyUpdate(key []byte, event interface{}) {
}

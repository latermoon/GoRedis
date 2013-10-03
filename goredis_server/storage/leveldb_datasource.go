package storage

import (
	"fmt"
	"github.com/syndtr/goleveldb/leveldb"
	"github.com/syndtr/goleveldb/leveldb/opt"
)

// 使用LevelDB做数据源
type LevelDBDataSource struct {
	DataSource
	db *leveldb.DB
	ro *opt.ReadOptions
	wo *opt.WriteOptions
}

func NewLevelDBDataSource(path string) (l *LevelDBDataSource, err error) {
	l = &LevelDBDataSource{}
	l.ro = &opt.ReadOptions{}
	l.wo = &opt.WriteOptions{}
	l.db, err = leveldb.OpenFile(path, &opt.Options{Flag: opt.OFCreateIfMissing})
	return
}

func (l *LevelDBDataSource) Get(key string) (entry Entry) {
	// l.mutex.Lock()
	// defer l.mutex.Unlock()
	bs, e1 := l.db.Get([]byte(key), l.ro)
	if e1 != nil || len(bs) == 0 {
		return
	}

	//fmt.Println("Get Type", bs, string(bs), entryType)
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
func (l *LevelDBDataSource) Set(key string, entry Entry) (err error) {
	// l.mutex.Lock()
	// defer l.mutex.Unlock()
	var bs []byte
	bs, err = entry.Encode()
	if err == nil {
		buf := make([]byte, len(bs)+1)
		copy(buf, []byte{byte(entry.Type())})
		copy(buf[1:], bs)
		err = l.db.Put([]byte(key), buf, l.wo)
	}
	return
}

func (l *LevelDBDataSource) Remove(key string) (err error) {
	// l.mutex.Lock()
	// defer l.mutex.Unlock()
	err = l.db.Delete([]byte(key), l.wo)
	return
}

func (l *LevelDBDataSource) NotifyEntryUpdate(key string, entry Entry) {
	go func() {
		l.Set(key, entry)
	}()
}

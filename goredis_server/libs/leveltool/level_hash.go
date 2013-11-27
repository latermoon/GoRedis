package leveltool

/*
__key[profile]hash = ""
__hash[profile]name = latermoon
__hash[profile]age = 27
__hash[profile]sex = M
*/

import (
	// "fmt"
	"bytes"
	"github.com/syndtr/goleveldb/leveldb"
	"github.com/syndtr/goleveldb/leveldb/iterator"
	"github.com/syndtr/goleveldb/leveldb/opt"
	"strings"
	"sync"
)

type HashElem struct {
	Key   []byte
	Value []byte
}

type LevelHash struct {
	db *leveldb.DB
	ro *opt.ReadOptions
	wo *opt.WriteOptions
	// key
	entryKey string
	mu       sync.Mutex
}

func NewLevelHash(db *leveldb.DB, entryKey string) (l *LevelHash) {
	l = &LevelHash{}
	l.db = db
	l.entryKey = entryKey
	l.ro = &opt.ReadOptions{}
	l.wo = &opt.WriteOptions{}
	return
}

func (l *LevelHash) Size() int {
	return 0
}

func (l *LevelHash) infoKey() []byte {
	return []byte(strings.Join([]string{KEY_PREFIX, SEP_LEFT, l.entryKey, SEP_RIGHT, HASH_SUFFIX}, ""))
}

func (l *LevelHash) infoValue() []byte {
	return []byte{}
}

func (l *LevelHash) fieldKey(field []byte) []byte {
	return bytes.Join([][]byte{l.fieldPrefix(), field}, []byte{})
}

func (l *LevelHash) fieldPrefix() []byte {
	return []byte(strings.Join([]string{HASH_PREFIX, SEP_LEFT, l.entryKey, SEP_RIGHT}, ""))
}

// 从fieldkey中提取field
func (l *LevelHash) fieldInKey(fieldkey []byte) (field []byte) {
	right := bytes.Index(fieldkey, []byte(SEP_RIGHT))
	return copyBytes(fieldkey[right+1:])
}

func (l *LevelHash) Get(field []byte) (val []byte) {
	fieldkey := l.fieldKey(field)
	var err error
	val, err = l.db.Get(fieldkey, l.ro)
	if err != nil {
		val = nil
	}
	return
}

func (l *LevelHash) Set(fieldVals ...[]byte) (n int) {
	l.mu.Lock()
	defer l.mu.Unlock()
	batch := new(leveldb.Batch)
	n = 0
	for i := 0; i < len(fieldVals); i += 2 {
		fieldkey := l.fieldKey(fieldVals[i])
		val := fieldVals[i+1]
		batch.Put(fieldkey, val)
		n++
	}
	batch.Put(l.infoKey(), l.infoValue())
	l.db.Write(batch, l.wo)
	return
}

func (l *LevelHash) GetAll(limit int) (elems []*HashElem) {
	iter := l.db.NewIterator(l.ro)
	defer iter.Release()
	elems = make([]*HashElem, 0, 10)
	PrefixEnumerate(iter, l.fieldPrefix(), func(i int, iter iterator.Iterator, quit *bool) {
		if limit != -1 && i >= limit {
			*quit = true
			return
		}
		elem := &HashElem{}
		elem.Key = l.fieldInKey(iter.Key())
		elem.Value = copyBytes(iter.Value())
		elems = append(elems, elem)
	}, "next")
	return
}

func (l *LevelHash) Exist(field []byte) (exist bool) {
	val := l.Get(field)
	exist = val != nil
	return
}

func (l *LevelHash) Remove(fields ...[]byte) (n int) {
	l.mu.Lock()
	defer l.mu.Unlock()
	n = 0
	for _, field := range fields {
		if l.Get(field) != nil {
			l.db.Delete(l.fieldKey(field), l.wo)
			n++
		}
	}
	// 检查是否已经删除完
	hasElem := false
	iter := l.db.NewIterator(l.ro)
	defer iter.Release()
	PrefixEnumerate(iter, l.fieldPrefix(), func(i int, iter iterator.Iterator, quit *bool) {
		hasElem = true
		*quit = true
	}, "next")
	if !hasElem {
		l.db.Delete(l.infoKey(), l.wo)
	}
	return
}

// 为了数据管理方便，这里不持久化count，每次都是枚举实现
// 为了性能保障，对于大于1000返回-1，不再扫描
func (l *LevelHash) Count() (n int) {
	iter := l.db.NewIterator(l.ro)
	defer iter.Release()
	n = 0
	PrefixEnumerate(iter, l.fieldPrefix(), func(i int, iter iterator.Iterator, quit *bool) {
		n++
		if n > 1000 {
			n = -1
			*quit = true
			return
		}
	}, "next")
	return
}

func (l *LevelHash) Drop() (ok bool) {
	iter := l.db.NewIterator(l.ro)
	defer iter.Release()
	batch := new(leveldb.Batch)
	PrefixEnumerate(iter, l.fieldPrefix(), func(i int, iter iterator.Iterator, quit *bool) {
		batch.Delete(copyBytes(iter.Key()))
	}, "next")
	batch.Delete(l.infoKey())
	l.db.Write(batch, l.wo)
	ok = true
	return
}

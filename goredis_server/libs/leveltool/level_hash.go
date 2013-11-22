package leveltool

/*
prefix:field:name = latermoon
prefix:field:age = 27
prefix:field:sex = M
*/

import (
	// "fmt"
	"github.com/syndtr/goleveldb/leveldb"
	"github.com/syndtr/goleveldb/leveldb/iterator"
	"github.com/syndtr/goleveldb/leveldb/opt"
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
	// key前缀
	prefix string
	mu     sync.Mutex
}

func NewLevelHash(db *leveldb.DB, prefix string) (l *LevelHash) {
	l = &LevelHash{}
	l.db = db
	l.prefix = prefix
	l.ro = &opt.ReadOptions{}
	l.wo = &opt.WriteOptions{}
	return
}

func (l *LevelHash) Get(field []byte) (val []byte) {
	fieldkey := l.fieldKey(field)
	var err error
	val, err = l.db.Get([]byte(fieldkey), l.ro)
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
		batch.Put([]byte(fieldkey), val)
		n++
	}
	l.db.Write(batch, l.wo)
	return
}

func (l *LevelHash) GetAll(limit int) (elems []*HashElem) {
	iter := l.db.NewIterator(l.ro)
	defer iter.Release()
	elems = make([]*HashElem, 0, 10)
	prefix := []byte(l.prefix + ":field:")
	PrefixEnumerate(iter, prefix, func(i int, iter iterator.Iterator, quit *bool) {
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
			l.db.Delete([]byte(l.fieldKey(field)), l.wo)
			n++
		}
	}
	return
}

// 为了数据管理方便，这里不持久化count，每次都是枚举实现
// 为了性能保障，对于大于1000返回-1，不再扫描
func (l *LevelHash) Count() (n int) {
	iter := l.db.NewIterator(l.ro)
	defer iter.Release()
	prefix := []byte(l.prefix + ":field:")
	n = 0
	PrefixEnumerate(iter, prefix, func(i int, iter iterator.Iterator, quit *bool) {
		n++
		if n > 1000 {
			n = -1
			*quit = true
			return
		}
	}, "next")
	return
}

// ===================================
func (l *LevelHash) fieldKey(field []byte) string {
	return l.prefix + ":field:" + string(field)
}

// 从fieldkey中提取field
func (l *LevelHash) fieldInKey(fieldkey []byte) (field []byte) {
	return copyBytes(fieldkey[len(l.prefix+":field:"):])
}

// ===================================

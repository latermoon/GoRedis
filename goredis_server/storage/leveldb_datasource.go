package storage

import (
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

func (l *LevelDBDataSource) Get(key string) (entry Entry, exist bool) {
	// l.mutex.Lock()
	// defer l.mutex.Unlock()
	bs, e1 := l.db.Get([]byte(key), l.ro)
	if e1 == nil {
		entry = NewStringEntry(bs)
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
	switch entry.Value().(type) {
	case string:
		bs = []byte(entry.Value().(string))
	default:
		bs = entry.Value().([]byte)
	}
	err = l.db.Put([]byte(key), bs, l.wo)
	return
}

func (l *LevelDBDataSource) Remove(key string) (err error) {
	// l.mutex.Lock()
	// defer l.mutex.Unlock()
	err = l.db.Delete([]byte(key), l.wo)
	return
}

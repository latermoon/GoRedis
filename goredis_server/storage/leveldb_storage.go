package storage

import (
	"github.com/syndtr/goleveldb/leveldb"
	"github.com/syndtr/goleveldb/leveldb/opt"
)

type LevelDBStorage struct {
	db *leveldb.DB
	ro *opt.ReadOptions
	wo *opt.WriteOptions
}

func NewLevelDBStringStorage(dbpath string) (storage *LevelDBStorage, err error) {
	storage = &LevelDBStorage{}
	storage.ro = &opt.ReadOptions{}
	storage.wo = &opt.WriteOptions{}
	storage.db, err = leveldb.OpenFile(dbpath, &opt.Options{Flag: opt.OFCreateIfMissing})
	return
}

func (s *LevelDBStorage) Get(key string) (value interface{}, err error) {
	data, e1 := s.db.Get([]byte(key), s.ro)
	if e1 == nil {
		value = data
	} else {
		value = nil
	}
	return
}

func (s *LevelDBStorage) Set(key string, value string) (err error) {
	err = s.db.Put([]byte(key), []byte(value), s.wo)
	return
}

func (s *LevelDBStorage) MGet(keys ...string) (values []interface{}, err error) {
	values = make([]interface{}, len(keys))
	for i, key := range keys {
		data, e1 := s.db.Get([]byte(key), s.ro)
		if e1 == nil {
			values[i] = data
		} else {
			values[i] = nil
		}
	}
	return
}

func (s *LevelDBStorage) MSet(keyvals ...string) (err error) {
	batch := new(leveldb.Batch)
	count := len(keyvals)
	for i := 0; i < count; i += 2 {
		batch.Put([]byte(keyvals[i]), []byte(keyvals[i+1]))
	}
	err = s.db.Write(batch, s.wo)
	return
}

func (s *LevelDBStorage) Del(keys ...string) (n int, err error) {
	n = len(keys)
	batch := new(leveldb.Batch)
	for _, key := range keys {
		batch.Delete([]byte(key))
	}
	err = s.db.Write(batch, s.wo)
	return
}

package storage

import (
	"github.com/syndtr/goleveldb/leveldb"
)

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

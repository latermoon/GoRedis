package storage

import (
	"github.com/syndtr/goleveldb/leveldb"
	"github.com/syndtr/goleveldb/leveldb/opt"
)

type LevelDBStorage struct {
	MemoryStorage
	db *leveldb.DB
	ro *opt.ReadOptions
	wo *opt.WriteOptions
}

func NewLevelDBStorage(dbpath string) (storage *LevelDBStorage, err error) {
	storage = &LevelDBStorage{}
	storage.ro = &opt.ReadOptions{}
	storage.wo = &opt.WriteOptions{}
	storage.db, err = leveldb.OpenFile(dbpath, &opt.Options{Flag: opt.OFCreateIfMissing})
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

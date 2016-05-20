package rocks

import (
	"github.com/tecbot/gorocksdb"
)

type DB struct {
	rdb *gorocksdb.DB
	wo  *gorocksdb.WriteOptions
	ro  *gorocksdb.ReadOptions
}

func New(rdb *gorocksdb.DB) *DB {
	db := &DB{rdb: rdb}
	db.wo = gorocksdb.NewDefaultWriteOptions()
	db.ro = gorocksdb.NewDefaultReadOptions()
	return db
}

func (d *DB) RawGet(key []byte) ([]byte, error) {
	return d.rdb.GetBytes(d.ro, key)
}

func (d *DB) RawSet(key, value []byte) error {
	return d.rdb.Put(d.wo, key, value)
}

func (d *DB) RawDelete(key []byte) error {
	return d.rdb.Delete(d.wo, key)
}

func (d *DB) Close() {
	d.wo.Destroy()
	d.ro.Destroy()
	d.rdb.Close()
}

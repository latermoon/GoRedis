package rocks

import (
	"bytes"
	"github.com/tecbot/gorocksdb"
)

type IterDirection int

const (
	IterForward IterDirection = iota
	IterBackward
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

func (d *DB) Hash(key []byte) *HashElement {
	return NewHashElement(d, key)
}

func (d *DB) Delete(key []byte) error {
	return nil
}

func (d *DB) Get(key []byte) ([]byte, error) {
	return d.RawGet(rawKey(key, STRING))
}

func (d *DB) Set(key, value []byte) error {
	return d.RawSet(rawKey(key, STRING), value)
}

func (d *DB) RawGet(key []byte) ([]byte, error) {
	return d.rdb.GetBytes(d.ro, key)
}

func (d *DB) RawSet(key, value []byte) error {
	return d.rdb.Put(d.wo, key, value)
}

func (d *DB) WriteBatch(batch *gorocksdb.WriteBatch) error {
	return d.rdb.Write(d.wo, batch)
}

func (d *DB) RawDelete(key []byte) error {
	return d.rdb.Delete(d.wo, key)
}

func (d *DB) Close() {
	d.wo.Destroy()
	d.ro.Destroy()
	d.rdb.Close()
}

// 前缀扫描
func (d *DB) PrefixEnumerate(prefix []byte, direction IterDirection, fn func(i int, key, value []byte, quit *bool)) {
	min := prefix
	max := append(prefix, MAXBYTE)
	j := -1
	d.RangeEnumerate(min, max, direction, func(i int, key, value []byte, quit *bool) {
		if bytes.HasPrefix(key, prefix) {
			j++
			fn(j, key, value, quit)
		} else {
			// 根据rocksdb的key有序，因此具有相同前缀的key必定是在一起的
			// 所以一旦碰见了没有该前缀的key那么就直接退出，结束遍历
			*quit = true
		}
	})
	return
}

func (d *DB) RangeEnumerate(min, max []byte, direction IterDirection, fn func(i int, key, value []byte, quit *bool)) {
	opts := gorocksdb.NewDefaultReadOptions()
	opts.SetFillCache(false)
	defer opts.Destroy()
	iter := d.rdb.NewIterator(opts)
	defer iter.Close()
	d.Enumerate(iter, min, max, direction, fn)
}

// 范围扫描
func (d *DB) Enumerate(iter *gorocksdb.Iterator, min, max []byte, direction IterDirection, fn func(i int, key, value []byte, quit *bool)) {
	found := false
	if direction == IterBackward {
		if len(max) == 0 {
			iter.SeekToLast()
		} else {
			iter.Seek(max)
		}
	} else {
		if len(min) == 0 {
			iter.SeekToFirst()
		} else {
			iter.Seek(min)
		}

	}
	found = iter.Valid()
	if !found {
		return
	}

	i := -1
	// 范围判断
	if found && between(iter.Key().Data(), min, max) {
		i++
		quit := false
		fn(i, iter.Key().Data(), iter.Value().Data(), &quit)
		if quit {
			return
		}
	}
	for {
		found = false
		if direction == IterBackward {
			iter.Prev()
		} else {
			iter.Next()
		}
		found = iter.Valid()
		if found && between(iter.Key().Data(), min, max) {
			i++
			quit := false
			fn(i, iter.Key().Data(), iter.Value().Data(), &quit)
			if quit {
				return
			}
		} else {
			break
		}
	}

	return
}
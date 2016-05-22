package rocks

import (
	"bytes"
	"github.com/golang/groupcache/lru"
	"github.com/tecbot/gorocksdb"
	"sync"
)

// rocks.DB provide "RedisLike's" rocksdb operators
type DB struct {
	rdb    *gorocksdb.DB
	wo     *gorocksdb.WriteOptions
	ro     *gorocksdb.ReadOptions
	mu     sync.Mutex
	caches *lru.Cache
}

func New(rdb *gorocksdb.DB) *DB {
	db := &DB{rdb: rdb}
	db.wo = gorocksdb.NewDefaultWriteOptions()
	db.ro = gorocksdb.NewDefaultReadOptions()
	db.caches = lru.New(1000)
	db.RawSet([]byte{MAXBYTE}, nil) // for Enumerator seek to last
	return db
}

func (d *DB) objFromCache(key []byte, e ElementType) interface{} {
	d.mu.Lock()
	defer d.mu.Unlock()

	skey := string(key)
	obj, ok := d.caches.Get(skey)
	if !ok {
		switch e {
		case HASH:
			obj = NewHashElement(d, key)
		case LIST:
			obj = NewListElement(d, key)
		case SORTEDSET:
			obj = NewSortedSetElement(d, key)
		}
		d.caches.Add(skey, obj)
	}
	return obj
}

func (d *DB) Hash(key []byte) *HashElement {
	return d.objFromCache(key, HASH).(*HashElement)
}

func (d *DB) List(key []byte) *ListElement {
	return d.objFromCache(key, LIST).(*ListElement)
}

func (d *DB) SortedSet(key []byte) *SortedSetElement {
	return d.objFromCache(key, SORTEDSET).(*SortedSetElement)
}

func (d *DB) Delete(key []byte) error {
	return nil
}

func (d *DB) TypeOf(key []byte) ElementType {
	c := ElementType(NONE)
	prefix := bytes.Join([][]byte{KEY, key, SEP}, nil)
	d.PrefixEnumerate(prefix, IterForward, func(i int, key, value []byte, quit *bool) {
		c = ElementType(key[len(prefix):][0])
		*quit = true
	})
	return c
}

func (d *DB) Get(key []byte) ([]byte, error) {
	return d.RawGet(rawKey(key, STRING))
}

func (d *DB) Set(key, value []byte) error {
	return d.RawSet(rawKey(key, STRING), value)
}

func (d *DB) WriteBatch(batch *gorocksdb.WriteBatch) error {
	return d.rdb.Write(d.wo, batch)
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

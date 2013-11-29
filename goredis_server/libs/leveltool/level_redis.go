package leveltool

/*
基于leveldb实现的redis持久化层
*/

import (
	lru "../lrucache"
	"bytes"
	// "fmt"
	"github.com/syndtr/goleveldb/leveldb"
	"github.com/syndtr/goleveldb/leveldb/opt"
	"sync"
)

type IteratorDirection int

const (
	IteratorForward IteratorDirection = iota
	IteratorBackward
)

type LevelRedis struct {
	db       *leveldb.DB
	ro       *opt.ReadOptions
	wo       *opt.WriteOptions
	lruCache *lru.LRUCache // LRU缓存层
	mu       sync.Mutex
	lstring  *LevelString
	lkey     *LevelKey
}

func NewLevelRedis(db *leveldb.DB) (l *LevelRedis) {
	l = &LevelRedis{}
	l.db = db
	l.ro = &opt.ReadOptions{}
	l.wo = &opt.WriteOptions{}
	l.lstring = NewLevelString(db)
	l.lruCache = lru.NewLRUCache(10000)
	return
}

func (l *LevelRedis) Strings() (s *LevelString) {
	return l.lstring
}

func (l *LevelRedis) objFromCache(key string, fn func() interface{}) (obj interface{}) {
	l.mu.Lock()
	defer l.mu.Unlock()
	var ok bool
	obj, ok = l.lruCache.Get(key)
	if !ok {
		obj = fn()
		l.lruCache.Set(key, obj.(lru.Value))
	}
	return
}

func (l *LevelRedis) GetList(key string) (lst *LevelList) {
	obj := l.objFromCache(key, func() interface{} {
		return NewLevelList(l.db, key)
	})
	return obj.(*LevelList)
}

func (l *LevelRedis) GetHash(key string) (h *LevelHash) {
	obj := l.objFromCache(key, func() interface{} {
		return NewLevelHash(l.db, key, false)
	})
	return obj.(*LevelHash)
}

func (l *LevelRedis) GetSet(key string) (s *LevelHash) {
	obj := l.objFromCache(key, func() interface{} {
		return NewLevelHash(l.db, key, true)
	})
	return obj.(*LevelHash)
}

func (l *LevelRedis) GetSortedSet(key string) (z *LevelZSet) {
	obj := l.objFromCache(key, func() interface{} {
		return NewLevelZSet(l, key)
	})
	return obj.(*LevelZSet)
}

func (l *LevelRedis) TypeOf(key []byte) (t string) {
	min := joinStringBytes(KEY_PREFIX, SEP_LEFT, string(key), SEP_RIGHT)
	max := min
	l.Enumerate(min, max, IteratorForward, func(i int, key, value []byte, quit *bool) {
		right := bytes.LastIndex(key, []byte(SEP_RIGHT))
		t = string(key[right+1:])
		*quit = true
	})
	if len(t) == 0 {
		t = "none"
	}
	return
}

func (l *LevelRedis) Delete(keys ...[]byte) (n int) {
	n = 0
	for _, keybytes := range keys {
		key := string(keybytes)
		t := l.TypeOf(keybytes)
		switch t {
		case "string":
			n += l.Strings().Delete(keybytes)
		case "hash":
			ok := l.GetHash(key).Drop()
			if ok {
				n++
			}
		case "set":
			ok := l.GetSet(key).Drop()
			if ok {
				n++
			}
		case "list":
			ok := l.GetList(key).Drop()
			if ok {
				n++
			}
		case "zset":
			ok := l.GetSortedSet(key).Drop()
			if ok {
				n++
			}
		default:
		}
		// ensure remove from lrucache
		l.lruCache.Delete(key)
	}
	return
}

func (l *LevelRedis) Keys(prefix []byte, fn func(i int, key, keytype []byte, quit *bool)) {
	min := joinStringBytes(KEY_PREFIX, SEP_LEFT, string(prefix))
	max := append(min, 254)
	l.Enumerate(min, max, IteratorForward, func(i int, key, value []byte, quit *bool) {
		left := bytes.Index(key, []byte(SEP_LEFT))
		right := bytes.LastIndex(key, []byte(SEP_RIGHT))
		fn(i, key[left+1:right], key[right+1:], quit)
		// fmt.Println(string(min), string(max), i, string(key), string(value), string(keytype), *quit)
	})
}

func (l *LevelRedis) Enumerate(min, max []byte, direction IteratorDirection, fn func(i int, key, value []byte, quit *bool)) {
	iter := l.db.NewIterator(l.ro)
	defer iter.Release()
	found := false
	if direction == IteratorBackward {
		found = iter.Seek(max)
	} else {
		found = iter.Seek(min)
	}
	i := -1
	if found {
		i++
		quit := false
		fn(i, copyBytes(iter.Key()), copyBytes(iter.Value()), &quit)
		if quit {
			return
		}
	}
	for {
		found = false
		if direction == IteratorBackward {
			found = iter.Prev()
			if !found || bytes.Compare(iter.Key(), min) < 0 {
				break
			}
		} else {
			found = iter.Next()
			if !found || bytes.Compare(iter.Key(), max) > 0 {
				break
			}
		}
		i++
		quit := false
		fn(i, copyBytes(iter.Key()), copyBytes(iter.Value()), &quit)
		if quit {
			return
		}
	}

	return
}

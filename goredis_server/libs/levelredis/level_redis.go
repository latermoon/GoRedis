package levelredis

/*
基于leveldb实现的redis持久化层

string
	+[name]string = "latermoon"
	+[name]string#e1083 = "latermoon"
hash
	+[info]hash = ""
	_h[info]name = "latermoon"
	_h[info]age = "27"
	_h[info]sex = "M"
list
	+[list]list = "0,1"
	_l[list]#0 = "a"
	_l[list]#1 = "b"
	_l[list]#2 = "c"
	_l[list]#3 = "d"
zset
	+[user_rank]zset = "2"
	_z[user_rank]s#1002#100422 = ""
	_z[user_rank]s#1006#100423 = ""
	_z[user_rank]s#10102#300000 = ""
	_z[user_rank]m#100422 = "1002"
	_z[user_rank]m#100423 = "1006"
	_z[user_rank]m#300000 = "10102"
*/

import (
	lru "../lrucache"
	"bytes"
	// "fmt"
	"github.com/latermoon/levigo"
	"math"
	"sync"
)

// 共用字段
const (
	SEP        = "#"
	SEP_LEFT   = "["
	SEP_RIGHT  = "]"
	KEY_PREFIX = "+"
)

// 字节最大范围
const MAXBYTE byte = math.MaxUint8

// 数据结构的key后缀
const (
	STRING_SUFFIX = "string"
	HASH_SUFFIX   = "hash"
	LIST_SUFFIX   = "list"
	SET_SUFFIX    = "set"
	ZSET_SUFFIX   = "zset"
	DOC_SUFFIX    = "doc"
)

// 数据结构的key前缀
const (
	HASH_PREFIX = "_h"
	LIST_PREFIX = "_l"
	SET_PREFIX  = "_s"
	ZSET_PREFIX = "_z"
)

// 枚举方向
type IteratorDirection int

const (
	IteratorForward IteratorDirection = iota
	IteratorBackward
)

type LevelRedis struct {
	db       *levigo.DB
	ro       *levigo.ReadOptions
	wo       *levigo.WriteOptions
	lruCache *lru.LRUCache // LRU缓存层
	mu       sync.Mutex
	lstring  *LevelString
}

func NewLevelRedis(db *levigo.DB) (l *LevelRedis) {
	l = &LevelRedis{}
	l.db = db
	l.ro = levigo.NewReadOptions()
	l.wo = levigo.NewWriteOptions()
	l.lstring = NewLevelString(l)
	l.lruCache = lru.NewLRUCache(10000)
	return
}

func (l *LevelRedis) DB() (db *levigo.DB) {
	return l.db
}

func (l *LevelRedis) Strings() (s *LevelString) {
	return l.lstring
}

// 获取原始key的内容
func (l *LevelRedis) RawGet(key []byte) (value []byte) {
	value, _ = l.db.Get(l.ro, key)
	return
}

// 使用LRUCache管理string以外的数据结构实例
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
		return NewLevelList(l, key)
	})
	return obj.(*LevelList)
}

func (l *LevelRedis) GetHash(key string) (h *LevelHash) {
	obj := l.objFromCache(key, func() interface{} {
		return NewLevelHash(l, key, false)
	})
	return obj.(*LevelHash)
}

func (l *LevelRedis) GetSet(key string) (s *LevelHash) {
	obj := l.objFromCache(key, func() interface{} {
		return NewLevelHash(l, key, true)
	})
	return obj.(*LevelHash)
}

func (l *LevelRedis) GetSortedSet(key string) (z *LevelZSet) {
	obj := l.objFromCache(key, func() interface{} {
		return NewLevelZSet(l, key)
	})
	return obj.(*LevelZSet)
}

func (l *LevelRedis) GetDoc(key string) (d *LevelDocument) {
	obj := l.objFromCache(key, func() interface{} {
		return NewLevelDocument(l, key)
	})
	return obj.(*LevelDocument)
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
		case "doc":
			ok := l.GetDoc(key).Drop()
			if ok {
				n++
			}
		default:
		}
		if t != "string" {
			l.lruCache.Delete(key)
		}
	}
	return
}

// keys前缀扫描
func (l *LevelRedis) Keys(prefix []byte, fn func(i int, key, keytype []byte, quit *bool)) {
	rawprefix := joinStringBytes(KEY_PREFIX, SEP_LEFT, string(prefix))
	l.PrefixEnumerate(rawprefix, IteratorForward, func(i int, key, value []byte, quit *bool) {
		left := bytes.Index(key, []byte(SEP_LEFT))
		right := bytes.LastIndex(key, []byte(SEP_RIGHT))
		fn(i, key[left+1:right], key[right+1:], quit)
		// fmt.Println(string(min), string(max), i, string(key), string(value), string(keytype), *quit)
	})
}

func (l *LevelRedis) PrefixEnumerate(prefix []byte, direction IteratorDirection, fn func(i int, key, value []byte, quit *bool)) {
	iter := l.db.NewIterator(l.ro)
	defer iter.Close()
	found := false
	if direction == IteratorForward {
		iter.Seek(prefix)
	} else {
		var seek []byte
		// 从后面搜索时，设定适合的最大值
		if len(prefix) > 0 {
			seek = copyBytes(prefix)
			if prefix[len(prefix)-1] < MAXBYTE {
				// seek[len(seek)-1] = MAXBYTE
				seek = append(seek, MAXBYTE)
			}
		} else {
			seek = []byte{MAXBYTE}
		}
		// fmt.Println("seek ", string(prefix), " ", string(seek))
		iter.Seek(seek)
	}
	found = iter.Valid()
	if !found {
		return
	}

	i := -1
	if found && bytes.HasPrefix(iter.Key(), prefix) {
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
			iter.Prev()
			found = iter.Valid()
			if !found || !bytes.HasPrefix(iter.Key(), prefix) {
				break
			}
		} else {
			iter.Next()
			found = iter.Valid()
			if !found || !bytes.HasPrefix(iter.Key(), prefix) {
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

// key范围枚举
func (l *LevelRedis) Enumerate(min, max []byte, direction IteratorDirection, fn func(i int, key, value []byte, quit *bool)) {
	iter := l.db.NewIterator(l.ro)
	defer iter.Close()
	found := false
	if direction == IteratorBackward {
		iter.Seek(max)
	} else {
		iter.Seek(min)
	}
	found = iter.Valid()

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
			iter.Prev()
			found = iter.Valid()
			if !found || bytes.Compare(iter.Key(), min) < 0 {
				break
			}
		} else {
			iter.Next()
			found = iter.Valid()
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

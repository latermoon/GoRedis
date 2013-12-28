package levelredis

/*
基于leveldb实现的redis持久化层

1、key存储规则
为了提供keys、type等基本操作，每个存入的数据都会有这样的结构 +[key]type，用于表达key以及数据类型
比如一个set name latermoon，会在leveldb里产生 +[name]string = latermoon 的数据
对于string以外的复杂结构，还会有另外的字段，比如 hash 会有以_h开头的key，list会有_l开头的key

2、leveldb存储原则
因为整个设计都是为了海量存储的，所以所有支持的redis指令，都必须基于leveldb实现，不能消耗内存
必要的时候，会牺牲掉一些redis特性，比如list结构需要lindex的话，就必须放弃lrem和linsert

同时会对使用场景进行一些取舍，比如zset要提供zcard的话，就需要每次操作后更新len，但增加的一次leveldb操作会降低zadd性能
因此对于hash、set这种很少取count的数据，放弃hlen、scard的性能（但也可以提供1000以内的枚举统计）,来提高hset/sadd的性能

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
	+[user_rank]zset = ""
	_z[user_rank]m#100422 = "1002"
	_z[user_rank]m#100423 = "1006"
	_z[user_rank]m#300000 = "10102"
	_z[user_rank]s#1002#100422 = ""
	_z[user_rank]s#1006#100423 = ""
	_z[user_rank]s#10102#300000 = ""
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
	lruCache *lru.LRUCache // LRU缓存层，管理string以外的key
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
	// 初始化最大的key，对于Enumerate从后面开始扫描key非常重要
	// 使iter.Seek(key)必定Valid=true
	maxkey := []byte{MAXBYTE}
	l.RawSet(maxkey, nil)
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

func (l *LevelRedis) RawSet(key []byte, value []byte) {
	l.db.Put(l.wo, key, value)
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
		return NewLevelHash(l, key)
	})
	return obj.(*LevelHash)
}

func (l *LevelRedis) GetSet(key string) (s *LevelHash) {
	obj := l.objFromCache(key, func() interface{} {
		return NewLevelSet(l, key)
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
	prefix := joinStringBytes(KEY_PREFIX, SEP_LEFT, string(key), SEP_RIGHT)
	l.PrefixEnumerate(prefix, IteratorForward, func(i int, key, value []byte, quit *bool) {
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
	for _, keybytes := range keys {
		key := string(keybytes)
		t := l.TypeOf(keybytes)
		switch t {
		case "string":
			n += l.Strings().Delete(keybytes)
		case "hash":
			if ok := l.GetHash(key).Drop(); ok {
				n++
			}
		case "set":
			if ok := l.GetSet(key).Drop(); ok {
				n++
			}
		case "list":
			if ok := l.GetList(key).Drop(); ok {
				n++
			}
		case "zset":
			if ok := l.GetSortedSet(key).Drop(); ok {
				n++
			}
		case "doc":
			if ok := l.GetDoc(key).Drop(); ok {
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
	})
}

// 前缀扫描
func (l *LevelRedis) PrefixEnumerate(prefix []byte, direction IteratorDirection, fn func(i int, key, value []byte, quit *bool)) {
	min := prefix
	max := append(prefix, MAXBYTE)
	j := -1
	l.Enumerate(min, max, direction, func(i int, key, value []byte, quit *bool) {
		if bytes.HasPrefix(key, prefix) {
			j++
			fn(j, key, value, quit)
		}
	})
	return
}

// 范围扫描
func (l *LevelRedis) Enumerate(min, max []byte, direction IteratorDirection, fn func(i int, key, value []byte, quit *bool)) {
	iter := l.db.NewIterator(l.ro)
	defer iter.Close()

	found := false
	if direction == IteratorBackward {
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
	if found && between(iter.Key(), min, max) {
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
		} else {
			iter.Next()
		}
		found = iter.Valid()
		if found && between(iter.Key(), min, max) {
			i++
			quit := false
			fn(i, copyBytes(iter.Key()), copyBytes(iter.Value()), &quit)
			if quit {
				return
			}
		} else {
			break
		}
	}

	return
}

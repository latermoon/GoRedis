package levelredis

import (
	"GoRedis/libs/gorocks"
	lru "GoRedis/libs/lrucache"
	"bytes"
	"errors"
	"math"
	"sync"
)

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
hash
	+[info]hash = ""
	_h[info]name = "latermoon"
	_h[info]age = "27"
	_h[info]sex = "M"
list
	+[list]list = "0,3"
	_l[list]#0 = "a"
	_l[list]#1 = "b"
	_l[list]#2 = "c"
	_l[list]#3 = "d"
zset
	+[user_rank]zset = "3"
	_z[user_rank]m#100422 = "-2"
	_z[user_rank]m#100423 = "1"
	_z[user_rank]m#300000 = "2"
	_z[user_rank]s#-2#100422 = ""
	_z[user_rank]s#1#100423 = ""
	_z[user_rank]s#2#300000 = ""
*/

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
type IterDirection int

const (
	IterForward IterDirection = iota
	IterBackward
)

const (
	NoCompression     = gorocks.NoCompression
	SnappyCompression = gorocks.SnappyCompression
)

var (
	lruCacheSize         = uint64(10000) // cache size
	objCacheCreateThread = 100           // obj create threads
)

type LevelElem interface {
	Key() string
	Type() string
	Drop() bool
}

type LevelRedis struct {
	db       *gorocks.DB
	ro       *gorocks.ReadOptions
	wo       *gorocks.WriteOptions
	lruCache *lru.LRUCache // LRU缓存，管理string以外的key
	mus      []sync.Mutex  // Key Hash线程池
	lstring  *LevelString
	g        *global
	// stats
	muCount  sync.Mutex
	counters map[string]int64
	snap     *gorocks.Snapshot
	// closing func
	closing func()
}

// snapshot，快照模式
func NewLevelRedis(db *gorocks.DB, snapshot bool) (l *LevelRedis) {
	l = &LevelRedis{}
	l.db = db
	l.ro = gorocks.NewReadOptions()
	if snapshot {
		l.snap = l.db.NewSnapshot() //必须调用Close()释放snap
		l.ro.SetSnapshot(l.snap)
		l.ro.SetFillCache(false)
		l.wo = nil // snapshot模式下禁止写入数据
	} else {
		l.wo = gorocks.NewWriteOptions()
	}
	l.counters = map[string]int64{"get": 0, "set": 0, "batch": 0, "del": 0, "enum": 0, "lru_hit": 0, "lru_miss": 0}
	l.lstring = NewLevelString(l)
	l.g = newGlobal(l)
	l.lruCache = lru.NewLRUCache(lruCacheSize)
	l.mus = make([]sync.Mutex, objCacheCreateThread)
	// 初始化最大的key，对于Enumerate从后面开始扫描key非常重要
	// 使iter.Seek(key)必定Valid=true
	if !snapshot {
		maxkey := []byte{MAXBYTE}
		l.RawSet(maxkey, nil)
	}
	return
}

func (l *LevelRedis) Snapshot() *LevelRedis {
	return NewLevelRedis(l.db, true)
}

func (l *LevelRedis) DB() (db *gorocks.DB) {
	return l.db
}

func (l *LevelRedis) SetClosingFunc(fn func()) {
	l.closing = fn
}

func (l *LevelRedis) Close() {
	l.ro.Close()
	if l.wo != nil {
		l.wo.Close()
	}
	// 处于snap模式的话，db不属于自己，只关闭其他变量
	if l.snap != nil {
		l.db.ReleaseSnapshot(l.snap)
		l.snap = nil
	} else {
		l.db.Close()
	}
	if l.closing != nil {
		l.closing()
	}
}

func NewOptions() *gorocks.Options {
	return gorocks.NewOptions()
}

func NewDefaultEnv() *gorocks.Env {
	return gorocks.NewDefaultEnv()
}

func NewLRUCache(capacity int) *gorocks.Cache {
	return gorocks.NewLRUCache(capacity)
}

func Open(dbname string, o *gorocks.Options) (*gorocks.DB, error) {
	return gorocks.Open(dbname, o)
}

func Repair(dbname string) error {
	opts := gorocks.NewOptions()
	defer opts.Close()
	opts.SetCache(gorocks.NewLRUCache(128 * 1024 * 1024))
	opts.SetCompression(gorocks.SnappyCompression)
	opts.SetBlockSize(32 * 1024)
	opts.SetMaxBackgroundCompactions(6)
	opts.SetWriteBufferSize(128 * 1024 * 1024)
	opts.SetMaxOpenFiles(100000)
	opts.SetCreateIfMissing(true)
	env := gorocks.NewDefaultEnv()
	defer env.Close()
	env.SetBackgroundThreads(6)
	env.SetHighPriorityBackgroundThreads(2)
	opts.SetEnv(env)

	return gorocks.RepairDatabase(dbname, opts)
}

func (l *LevelRedis) Stats() string {
	return l.db.PropertyValue("rocksdb.stats")
}

func (l *LevelRedis) Global() *global {
	return l.g
}

// leveldb操作数，计数器
func (l *LevelRedis) incrCounter(name string) {
	l.muCount.Lock()
	l.counters[name]++
	l.muCount.Unlock()
}

func (l *LevelRedis) Counter(name string) int64 {
	return l.counters[name]
}

func (l *LevelRedis) Strings() (s *LevelString) {
	return l.lstring
}

// 获取原始key的内容
func (l *LevelRedis) RawGet(key []byte) (value []byte, err error) {
	l.incrCounter("get")
	value, err = l.db.Get(l.ro, key)
	return
}

func (l *LevelRedis) RawSet(key []byte, value []byte) error {
	if l.snap != nil {
		return errors.New("RawSet not allowed")
	}
	l.incrCounter("set")
	return l.db.Put(l.wo, key, value)
}

func (l *LevelRedis) RawDel(key []byte) error {
	if l.snap != nil {
		return errors.New("RawDel not allowed")
	}
	l.incrCounter("del")
	return l.db.Delete(l.wo, key)
}

func (l *LevelRedis) WriteBatch(w *gorocks.WriteBatch) error {
	if l.snap != nil {
		return errors.New("WriteBatch not allowed")
	}
	l.incrCounter("batch")
	return l.db.Write(l.wo, w)
}

// 使用LRUCache管理string以外的数据结构实例
func (l *LevelRedis) objFromCache(key string, fn func() interface{}) (obj interface{}) {
	// 因为level对象构造需要时间，这里使用多个mutex来多线程处理，同一个key只会hash到同一个mutex里
	mu := l.mus[SumOfStringChars(key)%objCacheCreateThread]
	mu.Lock()
	defer mu.Unlock()

	var ok bool
	obj, ok = l.lruCache.Get(key)
	if !ok {
		obj = fn()
		l.lruCache.Set(key, obj.(lru.Value))
		l.incrCounter("lru_miss")
	} else {
		l.incrCounter("lru_hit")
	}
	return
}

func (l *LevelRedis) GetElem(key string, typ string) (e LevelElem) {
	switch typ {
	case STRING_SUFFIX:
	case LIST_SUFFIX:
		return l.GetList(key)
	case HASH_SUFFIX:
		return l.GetHash(key)
	case SET_SUFFIX:
		return l.GetSet(key)
	case ZSET_SUFFIX:
		return l.GetSortedSet(key)
	default:
		e = nil
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

func (l *LevelRedis) GetDoc(key string) (d *LevelDoc) {
	obj := l.objFromCache(key, func() interface{} {
		return NewLevelDoc(l, key)
	})
	return obj.(*LevelDoc)
}

func (l *LevelRedis) TypeOf(key []byte) (t string) {
	prefix := joinStringBytes(KEY_PREFIX, SEP_LEFT, string(key), SEP_RIGHT)
	l.PrefixEnumerate(prefix, IterForward, func(i int, key, value []byte, quit *bool) {
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

		if t == "string" {
			n += l.Strings().Delete(keybytes)
		} else if t == "none" {
			continue
		} else {
			// 使用相同的lock来处理对象的创建和删除
			mu := l.mus[SumOfStringChars(key)%objCacheCreateThread]
			mu.Lock()
			defer mu.Unlock()

			elem := l.GetElem(key, t)
			if elem != nil {
				if ok := elem.Drop(); ok {
					n++
				}
			}
			l.lruCache.Delete(key)
		}
	}
	return
}

// keys前缀扫描
func (l *LevelRedis) Keys(prefix []byte, fn func(i int, key, keytype []byte, quit *bool)) {
	rawprefix := joinStringBytes(KEY_PREFIX, SEP_LEFT, string(prefix))
	l.PrefixEnumerate(rawprefix, IterForward, func(i int, key, value []byte, quit *bool) {
		left := bytes.Index(key, []byte(SEP_LEFT))
		right := bytes.LastIndex(key, []byte(SEP_RIGHT))
		fn(i, key[left+1:right], key[right+1:], quit)
	})
}

// 前缀扫描
func (l *LevelRedis) PrefixEnumerate(prefix []byte, direction IterDirection, fn func(i int, key, value []byte, quit *bool)) {
	min := prefix
	max := append(prefix, MAXBYTE)
	j := -1
	l.RangeEnumerate(min, max, direction, func(i int, key, value []byte, quit *bool) {
		if bytes.HasPrefix(key, prefix) {
			j++
			fn(j, key, value, quit)
		}
	})
	return
}

// key顺序扫描，常用于数据导出、附近搜索
// 返回的key是面向用户的key，而非内部结构的raw_key
func (l *LevelRedis) KeyEnumerate(seek []byte, direction IterDirection, fn func(i int, key, keytype, value []byte, quit *bool)) {
	var iter *gorocks.Iterator
	if l.snap != nil {
		iter = l.db.NewIterator(l.ro)
	} else {
		ro := gorocks.NewReadOptions()
		ro.SetFillCache(false)
		defer ro.Close()
		iter = l.db.NewIterator(ro)
	}
	defer iter.Close()

	minkey := joinStringBytes(KEY_PREFIX, SEP_LEFT, string(seek))
	maxkey := []byte{MAXBYTE}
	l.Enumerate(iter, minkey, maxkey, direction, func(i int, key, value []byte, quit *bool) {
		if !bytes.HasPrefix(key, minkey) {
			*quit = true
			return
		}
		left := bytes.Index(key, []byte(SEP_LEFT))
		right := bytes.LastIndex(key, []byte(SEP_RIGHT))
		if left == -1 || right == -1 {
			return // just skip
		}
		fn(i, key[left+1:right], key[right+1:], value, quit)
	})
}

func (l *LevelRedis) RangeEnumerate(min, max []byte, direction IterDirection, fn func(i int, key, value []byte, quit *bool)) {
	var iter *gorocks.Iterator
	if l.snap != nil {
		iter = l.db.NewIterator(l.ro)
	} else {
		ro := gorocks.NewReadOptions()
		ro.SetFillCache(false)
		defer ro.Close()
		iter = l.db.NewIterator(ro)
	}
	defer iter.Close()
	l.Enumerate(iter, min, max, direction, fn)
}

// 范围扫描
func (l *LevelRedis) Enumerate(iter *gorocks.Iterator, min, max []byte, direction IterDirection, fn func(i int, key, value []byte, quit *bool)) {
	l.incrCounter("enum")
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
		if direction == IterBackward {
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

package goredis_server

import (
	. "../goredis"
	"./libs/leveltool"
	lru "./libs/lrucache"
	. "./storage"
	"fmt"
	"github.com/syndtr/goleveldb/leveldb"
	"sync"
	"time"
)

// 带缓存的数据源，解决更新复杂数据的性能问题
// 对string直接操作leveldb，其他数据类型改为异步aof更新
type BufferDataSource struct {
	DataSource
	ldb   *LevelDBDataSource   // leveldb持久化层
	cache *lru.LRUCache        // LRU缓存层
	aof   *leveltool.LevelList // AOF队列
	mutex sync.Mutex
}

func NewBufferDataSource(ldb *LevelDBDataSource) (ds *BufferDataSource) {
	ds = &BufferDataSource{}
	ds.ldb = ldb
	ds.cache = lru.NewLRUCache(100000)
	ds.aof = leveltool.NewLevelList(ldb.DB(), "__goredis:cmdaof")
	go ds.aofSaveRunloop()
	return
}

// 将aof保持到ldb
func (ds *BufferDataSource) aofSaveRunloop() {
	for {
		if ds.aof.Len() == 0 {
			time.Sleep(time.Millisecond * 100)
			continue
		}
		elem, _ := ds.aof.Pop()
		bs := elem.Value.([]byte)
		fmt.Println("aof pop", string(bs))
	}
}

func (ds *BufferDataSource) DB() *leveldb.DB {
	return ds.ldb.DB()
}

func (ds *BufferDataSource) Get(key []byte) (entry Entry) {
	val, ok := ds.cache.Get(string(key))
	if ok {
		entry = val.(Entry)
	} else {
		entry = ds.ldb.Get(key)
		// 将string以外的数据结构缓存起来
		if entry != nil && entry.Type() != EntryTypeString {
			ds.cache.Set(string(key), entry)
		}
	}
	return
}

func (ds *BufferDataSource) Set(key []byte, entry Entry) (err error) {
	// 实时保存string
	if entry.Type() == EntryTypeString {
		err = ds.ldb.Set(key, entry)
	} else {
		ds.cache.Set(string(key), entry)
	}
	return
}

func (ds *BufferDataSource) Keys(pattern string) (keys []string) {
	keys = ds.ldb.Keys(pattern)
	return
}

func (ds *BufferDataSource) Remove(key []byte) (err error) {
	ds.cache.Delete(string(key))
	ds.ldb.Remove(key)
	return
}

func (ds *BufferDataSource) NotifyUpdate(key []byte, event interface{}) {
	cmd := event.(*Command)
	ds.aof.Push(cmd.Bytes())
}

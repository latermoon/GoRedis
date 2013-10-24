package goredis_server

import (
	. "../goredis"
	"./libs/codec"
	"./libs/leveltool"
	lru "./libs/lrucache"
	. "./storage"
	"github.com/syndtr/goleveldb/leveldb"
	"strings"
	"sync"
	"time"
)

var (
	mh = codec.MsgpackHandle{}
)

// 带缓存的数据源，解决更新复杂数据的性能问题
// 对string直接操作leveldb，其他数据类型改为异步aof更新
type BufferDataSource struct {
	DataSource
	ldb        *LevelDBDataSource   // leveldb持久化层
	cache      *lru.LRUCache        // LRU缓存层
	aof        *leveltool.LevelList // AOF队列
	mergeCache map[string]Entry     // 合并日志时用的暂存区
	mutex      sync.Mutex
}

func NewBufferDataSource(ldb *LevelDBDataSource) (ds *BufferDataSource) {
	ds = &BufferDataSource{}
	ds.ldb = ldb
	ds.cache = lru.NewLRUCache(100000)
	ds.aof = leveltool.NewLevelList(ldb.DB(), "__goredis:cmdaof")
	// go ds.aofSaveRunloop()
	return
}

// 检查上次退出未保存的日志数据，并持久化
func (ds *BufferDataSource) CheckUnsavedLog() {
	count := ds.aof.Len()
	ds.mergeLog(int(count))
}

// 将log中的command合并到数据库
func (ds *BufferDataSource) mergeLog(count int) {
	cache := make(map[string]Entry)
	for i := 0; i < count; i++ {
		elem, _ := ds.aof.Pop()
		if elem == nil {
			break
		}
		bs := elem.Value.([]byte)
		cmd, err := ds.decodeCommand(bs)
		if err != nil {
			stdlog.Error("[BufferDataSource] bad cmd %s, %s", cmd, err)
		} else {
			stdlog.Info("[BufferDataSource] save log <%s>", cmd)
		}
		cmdName := strings.ToUpper(cmd.Name())
		switch cmdName {
		case "DEL":
			keys := cmd.Args[1:]
			for _, key := range keys {
				ds.ldb.Remove(key)
			}
		case "HSET":
			key, _ := cmd.ArgAtIndex(1)
			field := cmd.StringAtIndex(2)
			value, _ := cmd.ArgAtIndex(3)
		case "HMSET":
		case "HDEL":
		default:
			stdlog.Warn("[BufferDataSource] ignore <%s>", cmd)
		}
	}
}

func (ds *BufferDataSource) getEntryFromCache(key []byte, et EntryType) (entry Entry) {
	var exist bool
	entry, exist = ds.mergeCache[string(key)]
	if !exist {
		switch et {
		case EntryTypeString:
		case EntryTypeHash:
		case EntryTypeList:
		case EntryTypeSet:
		case EntryTypeSortedSet:
		}
	}
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
		cmd, err := ds.decodeCommand(bs)
		if err != nil {
			stdlog.Error("[BufferDataSource] bad cmd %s, %s", cmd, err)
		} else {
			stdlog.Info("[BufferDataSource] save <%s>", cmd)
		}
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

// 实时保存string，其余只保持在内存，string以外的更新，必须依赖NotifyUpdate(...)
func (ds *BufferDataSource) Set(key []byte, entry Entry) (err error) {
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
	bs, err := ds.encodeCommand(cmd)
	if err != nil {
		stdlog.Error("[BufferDataSource] encode err %s, %s", cmd, err)
	} else {
		ds.aof.Push(bs)
	}
}

// ==========
func (ds *BufferDataSource) encodeCommand(cmd *Command) (bs []byte, err error) {
	enc := codec.NewEncoderBytes(&bs, &mh)
	err = enc.Encode(cmd.Args)
	return
}

func (ds *BufferDataSource) decodeCommand(bs []byte) (cmd *Command, err error) {
	dec := codec.NewDecoderBytes(bs, &mh)
	var args [][]byte
	err = dec.Decode(&args)
	if err == nil {
		cmd = NewCommand(args...)
	}
	return
}

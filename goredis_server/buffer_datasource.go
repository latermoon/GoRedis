package goredis_server

import (
	. "../goredis"
	"./libs/codec"
	"./libs/leveltool"
	lru "./libs/lrucache"
	. "./storage"
	"errors"
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
	ds.mergeCache = make(map[string]Entry)
	// go ds.aofSaveRunloop()
	return
}

// 检查上次退出未保存的日志数据，并持久化
func (ds *BufferDataSource) CheckUnsavedLog() {
	count := ds.aof.Len()
	if count > 0 {
		ds.mergeLog(int(count))
	}
}

// 将log中的command合并到数据库
func (ds *BufferDataSource) mergeLog(count int) {
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
			stdlog.Info("[BufferDataSource] merge log <%s>", cmd)
		}
		cmdName := strings.ToUpper(cmd.Name())
		switch cmdName {
		case "DEL":
			keys := cmd.Args[1:]
			for _, key := range keys {
				delete(ds.mergeCache, string(key))
				ds.ldb.Remove(key)
			}
		case "HSET":
			key := cmd.StringAtIndex(1)
			entry, e1 := ds.getEntryFromTmpCache(key, EntryTypeHash)
			if e1 != nil {
				stdlog.Warn("%s %s", e1, key)
				continue
			}
			field := cmd.StringAtIndex(2)
			value, _ := cmd.ArgAtIndex(3)
			entry.(*HashEntry).Set(field, value)
		case "HMSET":
			key := cmd.StringAtIndex(1)
			entry, e1 := ds.getEntryFromTmpCache(key, EntryTypeHash)
			if e1 != nil {
				stdlog.Warn("%s %s", e1, key)
				continue
			}
			keyvals := cmd.Args[2:]
			for i := 0; i < len(keyvals); i += 2 {
				field := string(keyvals[i])
				val := keyvals[i+1]
				entry.(*HashEntry).Set(field, val)
			}
		case "HDEL":
			key := cmd.StringAtIndex(1)
			entry, e1 := ds.getEntryFromTmpCache(key, EntryTypeHash)
			if e1 != nil {
				stdlog.Warn("%s %s", e1, key)
				continue
			}
			fields := cmd.StringArgs()[2:]
			n := 0
			for _, field := range fields {
				_, exist := entry.(*HashEntry).Map()[field]
				if exist {
					delete(entry.(*HashEntry).Map(), field)
					n++
				}
			}

			if len(entry.(*HashEntry).Map()) == 0 {
				delete(ds.mergeCache, key)
				ds.ldb.Remove([]byte(key))
			}
		default:
			stdlog.Warn("[BufferDataSource] ignore <%s>", cmd)
		}
	}
	// 序列化
	stdlog.Info("[BufferDataSource] save %d items, %d logs", len(ds.mergeCache), count)
	for key, entry := range ds.mergeCache {
		ds.ldb.Set([]byte(key), entry)
	}
	stdlog.Info("[BufferDataSource] done")
}

// 从临时缓存里获取，再从ldb中取，没有的话再创建
func (ds *BufferDataSource) getEntryFromTmpCache(key string, et EntryType) (entry Entry, err error) {
	var exist bool
	entry, exist = ds.mergeCache[key]
	if !exist {
		entry = ds.ldb.Get([]byte(key))
		// hash是异步的，set是同步的，如果一个key之前是hash，后来改为string，就会导致类型不一致
		if entry != nil && entry.Type() != et {
			err = errors.New("wrong entry type, might replace by string")
		} else if entry == nil {
			entry = NewEmptyEntry(et)
		}
		ds.mergeCache[key] = entry
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

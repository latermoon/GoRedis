package leveltool

/*
基于leveldb实现的list，主要用于海量存储，比如aof、日志

1、数据结构
[prefix]:_start = 1004 (int64)
[prefix]:_end = 1008 (int64)
[prefix]:idx:1004 = hello ([]byte)
[prefix]:idx:1005 = hello
[prefix]:idx:1006 = hello
[prefix]:idx:1007 = hello
[prefix]:idx:1008 = hello
*/

import (
	"encoding/binary"
	"github.com/syndtr/goleveldb/leveldb"
	"github.com/syndtr/goleveldb/leveldb/opt"
	"strconv"
	"sync"
)

type Element struct {
	Value interface{}
}

// LevelList的特点
// 类似双向链表，右进左出，可以通过索引查找
// 海量存储，占用内存小
type LevelList struct {
	db     *leveldb.DB
	ro     *opt.ReadOptions
	wo     *opt.WriteOptions
	prefix string
	start  int64
	end    int64
	mutex  sync.Mutex
}

func NewLevelList(db *leveldb.DB, prefix string) (lst *LevelList) {
	lst = &LevelList{}
	lst.db = db
	lst.ro = &opt.ReadOptions{}
	lst.wo = &opt.WriteOptions{}
	lst.prefix = prefix
	lst.start = lst.ldbGetInt64("_start", 0)
	lst.end = lst.ldbGetInt64("_end", -1)
	return
}

func (l *LevelList) ldbKey(key string) []byte {
	return []byte(l.prefix + ":" + key)
}

func (l *LevelList) startkey() []byte {
	return l.ldbKey("_start")
}

func (l *LevelList) endkey() []byte {
	return l.ldbKey("_end")
}

func (l *LevelList) idxkey(idx int64) []byte {
	return l.ldbKey("idx:" + strconv.FormatInt(idx, 10))
}

func (l *LevelList) ldbGetInt64(key string, defaultValue int64) int64 {
	data, err := l.db.Get(l.ldbKey(key), l.ro)
	if err != nil {
		return defaultValue
	}
	return bytesToInt64(data)
}

func (l *LevelList) ldbSetInt64(key string, value int64) (err error) {
	err = l.db.Put(l.ldbKey(key), int64ToBytes(value), l.wo)
	return
}

// func (l *LevelList) ldbSet(key string, value []byte) (err error) {
// 	err = l.db.Put(l.ldbKey(key), value, l.wo)
// 	return
// }

// 数据转换
func int64ToBytes(i int64) []byte {
	var buf = make([]byte, 8)
	binary.BigEndian.PutUint64(buf, uint64(i))
	return buf
}

// 数据转换
func bytesToInt64(buf []byte) int64 {
	return int64(binary.BigEndian.Uint64(buf))
}

func (l *LevelList) Push(value []byte) (e *Element, err error) {
	l.mutex.Lock()
	defer l.mutex.Unlock()

	idx := l.end + 1
	batch := new(leveldb.Batch)
	batch.Put(l.idxkey(idx), value)
	batch.Put(l.endkey(), int64ToBytes(idx))
	err = l.db.Write(batch, l.wo)
	if err == nil {
		// 写入成功后才移动游标
		l.end++
	}
	e = &Element{}
	e.Value = value
	return
}

func (l *LevelList) Pop() (e *Element, err error) {
	l.mutex.Lock()
	defer l.mutex.Unlock()
	if l.Len() == 0 {
		return nil, nil
	}
	idx := l.start
	e = &Element{}
	e.Value, err = l.db.Get(l.idxkey(idx), l.ro)
	if err != nil {
		return nil, err
	}
	// move, move, move
	batch := new(leveldb.Batch)
	batch.Put(l.startkey(), int64ToBytes(l.start+1))
	batch.Delete(l.idxkey(idx))
	err = l.db.Write(batch, l.wo)
	if err == nil {
		// 写入成功后才移动游标
		l.start++
	} else {
		return nil, err
	}
	return
}

func (l *LevelList) Element(idx int64) (e *Element) {
	if idx < 0 || idx >= l.Len() {
		return nil
	}
	value, err := l.db.Get(l.idxkey(idx), l.ro)
	if err != nil {
		return
	}
	e = &Element{}
	e.Value = value
	return
}

func (l *LevelList) Len() int64 {
	return l.end - l.start + 1
}

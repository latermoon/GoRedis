package leveltool

/*
基于leveldb实现的list，主要用于海量存储，比如aof、日志

1、数据结构
要提供序号访问，就不能删除中间的元素
[prefix]:_start = 1004 (int64)
[prefix]:_end = 1008 (int64)
[prefix]:idx:1004 = hello ([]byte)
[prefix]:idx:1005 = hello
[prefix]:idx:1006 = hello
[prefix]:idx:1007 = hello
[prefix]:idx:1008 = hello
*/
// 本页面命名注意，idx都表示大于l.start的那个索引序号，而不是0开始的数组序号

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
	db *leveldb.DB
	ro *opt.ReadOptions
	wo *opt.WriteOptions
	// key前缀
	prefix string
	// 游标控制
	start int64
	end   int64
	mutex sync.Mutex
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

// 获取配置里整形值
func (l *LevelList) ldbGetInt64(key string, defaultValue int64) int64 {
	data, err := l.db.Get(l.ldbKey(key), l.ro)
	if err != nil {
		return defaultValue
	}
	return bytesToInt64(data)
}

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

	// 添加数据并更新右游标
	idx := l.end + 1
	batch := new(leveldb.Batch)
	batch.Put(l.idxkey(idx), value)
	batch.Put(l.endkey(), int64ToBytes(idx))
	err = l.db.Write(batch, l.wo)
	if err == nil {
		// 写入成功
		l.end++
	}
	e = &Element{}
	e.Value = value
	return
}

func (l *LevelList) Pop() (e *Element, err error) {
	l.mutex.Lock()
	defer l.mutex.Unlock()

	if l.len() == 0 {
		return nil, nil
	}
	idx := l.start
	e = &Element{}
	e.Value, err = l.db.Get(l.idxkey(idx), l.ro)
	if err != nil {
		return nil, err
	}
	// 只剩下一个元素时，清除start、end
	shouldReset := l.len() == 1
	// 更新左游标，并删除数据
	batch := new(leveldb.Batch)
	if shouldReset {
		batch.Delete(l.startkey())
		batch.Delete(l.endkey())
	} else {
		batch.Put(l.startkey(), int64ToBytes(l.start+1))
	}
	batch.Delete(l.idxkey(idx))
	err = l.db.Write(batch, l.wo)
	if err == nil {
		// 删除成功
		if shouldReset {
			l.start = 0
			l.end = -1
		} else {
			l.start++
		}
	} else {
		return nil, err
	}
	return
}

func (l *LevelList) Index(i int64) (e *Element, err error) {
	l.mutex.Lock()
	defer l.mutex.Unlock()

	if i < 0 || i >= l.len() {
		return nil, nil
	}
	idx := l.start + i
	e = &Element{}
	e.Value, err = l.db.Get(l.idxkey(idx), l.ro)
	if err != nil {
		return nil, err
	}
	return
}

func (l *LevelList) len() int64 {
	return l.end - l.start + 1
}

func (l *LevelList) Len() int64 {
	l.mutex.Lock()
	defer l.mutex.Unlock()
	return l.len()
}

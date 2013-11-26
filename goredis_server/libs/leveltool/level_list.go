package leveltool

/*
基于leveldb实现的list，主要用于海量存储，比如aof、日志

1、数据结构
要提供序号访问，就不能删除中间的元素
__key[key]list = 1004,1008
__list[key]idx:1004 = hello
__list[key]idx:1005 = hello
__list[key]idx:1006 = hello
__list[key]idx:1007 = hello
__list[key]idx:1008 = hello
*/
// 本页面命名注意，idx都表示大于l.start的那个索引序号，而不是0开始的数组序号

import (
	"bytes"
	// "fmt"
	"github.com/syndtr/goleveldb/leveldb"
	"github.com/syndtr/goleveldb/leveldb/iterator"
	"github.com/syndtr/goleveldb/leveldb/opt"
	"strconv"
	"strings"
	"sync"
)

type Element struct {
	Value interface{}
}

// LevelList的特点
// 类似双向链表，右进左出，可以通过索引查找
// 海量存储，占用内存小
type LevelList struct {
	db       *leveldb.DB
	ro       *opt.ReadOptions
	wo       *opt.WriteOptions
	entryKey string
	// 游标控制
	start int64
	end   int64
	mu    sync.Mutex
}

func NewLevelList(db *leveldb.DB, entryKey string) (l *LevelList) {
	l = &LevelList{}
	l.db = db
	l.ro = &opt.ReadOptions{}
	l.wo = &opt.WriteOptions{}
	l.entryKey = entryKey
	l.start = 0
	l.end = -1
	l.initInfo()
	return
}

func (l *LevelList) Size() int {
	return 0
}

func (l *LevelList) initInfo() {
	data, err := l.db.Get(l.infoKey(), l.ro)
	if err != nil {
		return
	}
	pairs := bytes.Split(data, []byte(","))
	if len(pairs) < 2 {
		return
	}
	l.start, _ = strconv.ParseInt(string(pairs[0]), 10, 64)
	l.end, _ = strconv.ParseInt(string(pairs[1]), 10, 64)
}

// __key:[entry key]:list =
func (l *LevelList) infoKey() []byte {
	return []byte(strings.Join([]string{KEY_PREFIX, SEP_LEFT, l.entryKey, SEP_RIGHT, LIST_SUFFIX}, ""))
}

func (l *LevelList) infoValue() []byte {
	s := strconv.FormatInt(l.start, 10)
	e := strconv.FormatInt(l.end, 10)
	return []byte(s + "," + e)
}

func (l *LevelList) keyPrefix() []byte {
	return []byte(strings.Join([]string{LIST_PREFIX, SEP_LEFT, l.entryKey, SEP_RIGHT}, ""))
}

// __list:[key]:idx:1005 = hello
func (l *LevelList) idxKey(idx int64) []byte {
	idxStr := strconv.FormatInt(idx, 10)
	return []byte(strings.Join([]string{LIST_PREFIX, SEP_LEFT, l.entryKey, SEP_RIGHT, "idx", ":", idxStr}, ""))
}

func (l *LevelList) Push(value []byte) (e *Element, err error) {
	l.mu.Lock()
	defer l.mu.Unlock()

	// 添加数据并更新右游标
	l.end++
	batch := new(leveldb.Batch)
	batch.Put(l.idxKey(l.end), value)
	batch.Put(l.infoKey(), l.infoValue())
	err = l.db.Write(batch, l.wo)
	if err != nil {
		// 回退
		l.end--
	}
	e = &Element{}
	e.Value = value
	return
}

func (l *LevelList) Pop() (e *Element, err error) {
	l.mu.Lock()
	defer l.mu.Unlock()

	if l.len() == 0 {
		return nil, nil
	}
	// backup
	oldstart, oldend := l.start, l.end

	// get
	idx := l.start
	e = &Element{}
	e.Value, err = l.db.Get(l.idxKey(idx), l.ro)
	if err != nil {
		return nil, err
	}
	// 只剩下一个元素时，删除infoKey(0)
	shouldReset := l.len() == 1
	// 删除数据, 更新左游标
	batch := new(leveldb.Batch)
	batch.Delete(l.idxKey(idx))
	if shouldReset {
		l.start = 0
		l.end = -1
		batch.Delete(l.infoKey())
	} else {
		l.start++
		batch.Put(l.infoKey(), l.infoValue())
	}
	err = l.db.Write(batch, l.wo)
	if err != nil {
		// 回退
		l.start, l.end = oldstart, oldend
	}
	return
}

func (l *LevelList) Index(i int64) (e *Element, err error) {
	l.mu.Lock()
	defer l.mu.Unlock()

	if i < 0 || i >= l.len() {
		return nil, nil
	}
	idx := l.start + i
	e = &Element{}
	e.Value, err = l.db.Get(l.idxKey(idx), l.ro)
	if err != nil {
		return nil, err
	}
	return
}

func (l *LevelList) len() int64 {
	return l.end - l.start + 1
}

func (l *LevelList) Len() int64 {
	l.mu.Lock()
	defer l.mu.Unlock()
	return l.len()
}

func (l *LevelList) Drop() (n int) {
	iter := l.db.NewIterator(l.ro)
	defer iter.Release()
	batch := new(leveldb.Batch)
	PrefixEnumerate(iter, l.keyPrefix(), func(i int, iter iterator.Iterator, quit *bool) {
		batch.Delete(copyBytes(iter.Key()))
	}, "next")
	batch.Delete(l.infoKey())
	l.db.Write(batch, l.wo)
	n = 1
	return
}

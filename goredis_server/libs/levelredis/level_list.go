package levelredis

// 基于leveldb实现的list，主要用于海量存储，比如aof、日志
// 本页面命名注意，idx都表示大于l.start的那个索引序号，而不是0开始的数组序号

import (
	"bytes"
	// "fmt"
	"github.com/latermoon/levigo"
	"sync"
)

type Element struct {
	Value interface{}
}

// LevelList的特点
// 类似双向链表，右进左出，可以通过索引查找
// 海量存储，占用内存小
type LevelList struct {
	redis    *LevelRedis
	entryKey string
	// 游标控制
	start int64
	end   int64
	mu    sync.Mutex
}

func NewLevelList(redis *LevelRedis, entryKey string) (l *LevelList) {
	l = &LevelList{}
	l.redis = redis
	l.entryKey = entryKey
	l.start = l.firstIndex() // default 0
	l.end = l.lastIndex()    // default -1
	return
}

func (l *LevelList) Size() int {
	return 0
}

// __key:[entry key]:list =
func (l *LevelList) infoKey() []byte {
	return joinStringBytes(KEY_PREFIX, SEP_LEFT, l.entryKey, SEP_RIGHT, LIST_SUFFIX)
}

func (l *LevelList) infoValue() []byte {
	return []byte("")
}

func (l *LevelList) keyPrefix() []byte {
	return joinStringBytes(LIST_PREFIX, SEP_LEFT, l.entryKey, SEP_RIGHT)
}

// _l[key]#11005 = hello
func (l *LevelList) idxKey(idx int64) []byte {
	// 正负符号, 因为经过uint64转换后，负数的字典顺序比整数大，所以需要前置一个0、1保障顺序
	var sign string
	if idx < 0 {
		sign = "0"
	} else {
		sign = "1"
	}
	idxStr := string(Int64ToBytes(idx))
	return joinStringBytes(LIST_PREFIX, SEP_LEFT, l.entryKey, SEP_RIGHT, SEP, sign, idxStr)
}

func (l *LevelList) splitIndexKey(idxkey []byte) (idx int64) {
	pos := bytes.LastIndex(idxkey, []byte(SEP))
	idx = BytesToInt64(idxkey[pos+1+1:]) // +1 skip sign "0/1"
	return
}

func (l *LevelList) LPush(values ...[]byte) (err error) {
	l.mu.Lock()
	defer l.mu.Unlock()

	// 左游标
	oldstart := l.start
	oldlen := l.len()
	batch := levigo.NewWriteBatch()
	defer batch.Close()
	for _, value := range values {
		l.start--
		batch.Put(l.idxKey(l.start), value)
	}
	if oldlen == 0 { // 只需要第一次插入
		batch.Put(l.infoKey(), l.infoValue())
	}
	err = l.redis.db.Write(l.redis.wo, batch)
	if err != nil {
		// 回退
		l.start = oldstart
	}
	return
}

func (l *LevelList) RPush(values ...[]byte) (err error) {
	l.mu.Lock()
	defer l.mu.Unlock()

	// 右游标
	oldend := l.end
	oldlen := l.len()
	batch := levigo.NewWriteBatch()
	defer batch.Close()
	for _, value := range values {
		l.end++
		batch.Put(l.idxKey(l.end), value)
	}
	if oldlen == 0 { // 只需要第一次插入
		batch.Put(l.infoKey(), l.infoValue())
	}
	err = l.redis.db.Write(l.redis.wo, batch)
	if err != nil {
		// 回退
		l.end = oldend
	}
	return
}

func (l *LevelList) RPop() (e *Element, err error) {
	l.mu.Lock()
	defer l.mu.Unlock()

	if l.len() == 0 {
		return nil, nil
	}
	// backup
	oldstart, oldend := l.start, l.end

	// get
	idx := l.end
	e = &Element{}
	e.Value, err = l.redis.db.Get(l.redis.ro, l.idxKey(idx))
	if err != nil {
		return nil, err
	}

	// 只剩下一个元素时，删除infoKey(0)
	shouldReset := l.len() == 1
	// 删除数据, 更新左游标
	batch := levigo.NewWriteBatch()
	defer batch.Close()
	batch.Delete(l.idxKey(idx))
	if shouldReset {
		l.start = 0
		l.end = -1
		batch.Delete(l.infoKey())
	} else {
		l.end--
	}
	err = l.redis.db.Write(l.redis.wo, batch)
	if err != nil {
		// 回退
		l.start, l.end = oldstart, oldend
	}
	return
}

func (l *LevelList) LPop() (e *Element, err error) {
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
	e.Value, err = l.redis.db.Get(l.redis.ro, l.idxKey(idx))
	if err != nil {
		return nil, err
	}
	// 只剩下一个元素时，删除infoKey(0)
	shouldReset := l.len() == 1
	// 删除数据, 更新左游标
	batch := levigo.NewWriteBatch()
	defer batch.Close()
	batch.Delete(l.idxKey(idx))
	if shouldReset {
		l.start = 0
		l.end = -1
		batch.Delete(l.infoKey())
	} else {
		l.start++
	}
	err = l.redis.db.Write(l.redis.wo, batch)
	if err != nil {
		// 回退
		l.start, l.end = oldstart, oldend
	}
	return
}

// 删除左边
func (l *LevelList) TrimLeft(count uint) (n int) {
	l.mu.Lock()
	defer l.mu.Unlock()

	if l.len() == 0 {
		return
	}
	oldstart, oldend := l.start, l.end
	batch := levigo.NewWriteBatch()
	defer batch.Close()
	for idx := oldstart; idx < (oldstart+int64(count)) || idx <= oldend; idx++ {
		batch.Delete(l.idxKey(idx))
		l.start++
	}
	shouldReset := l.len() == 0
	if shouldReset {
		l.start = 0
		l.end = -1
		batch.Delete(l.infoKey())
	} else {
		batch.Put(l.infoKey(), l.infoValue())
	}
	err := l.redis.db.Write(l.redis.wo, batch)
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
	e.Value, err = l.redis.db.Get(l.redis.ro, l.idxKey(idx))
	if err != nil {
		return nil, err
	}
	return
}

func (l *LevelList) firstIndex() (n int64) {
	n = 0
	l.redis.PrefixEnumerate(l.keyPrefix(), IteratorForward, func(i int, key, value []byte, quit *bool) {
		n = l.splitIndexKey(key)
		*quit = true
	})
	return
}

func (l *LevelList) lastIndex() (n int64) {
	n = -1
	l.redis.PrefixEnumerate(l.keyPrefix(), IteratorBackward, func(i int, key, value []byte, quit *bool) {
		n = l.splitIndexKey(key)
		*quit = true
	})
	return
}

func (l *LevelList) len() int64 {
	if l.end < l.start {
		return 0
	}
	return l.end - l.start + 1
}

func (l *LevelList) Len() int64 {
	return l.len()
}

func (l *LevelList) Drop() (ok bool) {
	l.mu.Lock()
	defer l.mu.Unlock()
	batch := levigo.NewWriteBatch()
	defer batch.Close()
	l.redis.PrefixEnumerate(l.keyPrefix(), IteratorForward, func(i int, key, value []byte, quit *bool) {
		batch.Delete(key)
	})
	batch.Delete(l.infoKey())
	l.redis.db.Write(l.redis.wo, batch)
	ok = true
	l.start = 0
	l.end = -1
	return
}

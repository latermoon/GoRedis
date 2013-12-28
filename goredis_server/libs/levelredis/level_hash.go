package levelredis

import (
	// "fmt"
	"bytes"
	"github.com/latermoon/levigo"
	"sync"
)

type HashElem struct {
	Key   []byte
	Value []byte
}

// 使用userForSet控制实现set还是hash
type LevelHash struct {
	redis *LevelRedis
	// key
	entryKey string
	mu       sync.Mutex
	// for SET
	userForSet bool
}

// 构造方法1
func NewLevelSet(redis *LevelRedis, key string) (l *LevelHash) {
	l = &LevelHash{}
	l.redis = redis
	l.entryKey = key
	l.userForSet = true
	return
}

// 构造方法2
func NewLevelHash(redis *LevelRedis, entryKey string) (l *LevelHash) {
	l = &LevelHash{}
	l.redis = redis
	l.entryKey = entryKey
	l.userForSet = false
	return
}

func (l *LevelHash) Size() int {
	return 0
}

func (l *LevelHash) infoKey() []byte {
	if l.userForSet {
		return joinStringBytes(KEY_PREFIX, SEP_LEFT, l.entryKey, SEP_RIGHT, SET_SUFFIX)
	} else {
		return joinStringBytes(KEY_PREFIX, SEP_LEFT, l.entryKey, SEP_RIGHT, HASH_SUFFIX)
	}
}

func (l *LevelHash) infoValue() []byte {
	return []byte{}
}

func (l *LevelHash) fieldKey(field []byte) []byte {
	return bytes.Join([][]byte{l.fieldPrefix(), field}, []byte{})
}

func (l *LevelHash) fieldPrefix() []byte {
	if l.userForSet {
		return joinStringBytes(SET_PREFIX, SEP_LEFT, l.entryKey, SEP_RIGHT)
	} else {
		return joinStringBytes(HASH_PREFIX, SEP_LEFT, l.entryKey, SEP_RIGHT)
	}
}

// 从fieldkey中提取field
func (l *LevelHash) fieldInKey(fieldkey []byte) (field []byte) {
	right := bytes.Index(fieldkey, []byte(SEP_RIGHT))
	return copyBytes(fieldkey[right+1:])
}

func (l *LevelHash) Get(field []byte) (val []byte) {
	fieldkey := l.fieldKey(field)
	var err error
	val, err = l.redis.db.Get(l.redis.ro, fieldkey)
	if err != nil {
		val = nil
	}
	return
}

func (l *LevelHash) Set(fieldVals ...[]byte) (n int) {
	l.mu.Lock()
	defer l.mu.Unlock()

	batch := levigo.NewWriteBatch()
	defer batch.Close()
	n = 0
	for i := 0; i < len(fieldVals); i += 2 {
		fieldkey := l.fieldKey(fieldVals[i])
		val := fieldVals[i+1]
		batch.Put(fieldkey, val)
		n++
	}
	batch.Put(l.infoKey(), l.infoValue())
	l.redis.db.Write(l.redis.wo, batch)
	return
}

func (l *LevelHash) GetAll(limit int) (elems []*HashElem) {
	elems = make([]*HashElem, 0, 10)
	l.redis.PrefixEnumerate(l.fieldPrefix(), IteratorForward, func(i int, key, value []byte, quit *bool) {
		if limit != -1 && i >= limit {
			*quit = true
			return
		}
		elem := &HashElem{}
		elem.Key = l.fieldInKey(key)
		elem.Value = value
		elems = append(elems, elem)
	})
	return
}

func (l *LevelHash) Exist(field []byte) (exist bool) {
	val := l.Get(field)
	exist = val != nil
	return
}

func (l *LevelHash) Remove(fields ...[]byte) (n int) {
	l.mu.Lock()
	defer l.mu.Unlock()

	n = 0
	for _, field := range fields {
		if l.Get(field) != nil {
			l.redis.db.Delete(l.redis.wo, l.fieldKey(field))
			n++
		}
	}
	// 检查是否已经删除完
	hasElem := false
	l.redis.PrefixEnumerate(l.fieldPrefix(), IteratorForward, func(i int, key, value []byte, quit *bool) {
		hasElem = true
		*quit = true
	})
	if !hasElem {
		l.redis.db.Delete(l.redis.wo, l.infoKey())
	}
	return
}

// 为了数据管理方便，这里不持久化count，每次都是枚举实现
// 为了性能保障，对于大于100返回-1，不再扫描
func (l *LevelHash) Count() (n int) {
	l.redis.PrefixEnumerate(l.fieldPrefix(), IteratorForward, func(i int, key, value []byte, quit *bool) {
		n++
		if n > 100 {
			n = -1
			*quit = true
			return
		}
	})
	return
}

func (l *LevelHash) Drop() (ok bool) {
	l.mu.Lock()
	defer l.mu.Unlock()

	batch := levigo.NewWriteBatch()
	defer batch.Close()
	l.redis.PrefixEnumerate(l.fieldPrefix(), IteratorForward, func(i int, key, value []byte, quit *bool) {
		batch.Delete(key)
	})
	batch.Delete(l.infoKey())
	l.redis.db.Write(l.redis.wo, batch)
	ok = true
	return
}

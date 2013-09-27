package storage

import (
	"sync"
)

// 基于内存的Storage
type MemoryStorage struct {
	datamap map[string]interface{}
	mutex   *sync.Mutex
}

func NewMemoryStorage() (storage *MemoryStorage) {
	storage = &MemoryStorage{}
	storage.datamap = make(map[string]interface{})
	storage.mutex = &sync.Mutex{}
	return
}

// ============================================================
// 						String
// ============================================================

func (s *MemoryStorage) Get(key string) (value interface{}, err error) {
	value, _ = s.datamap[key]
	return
}

func (s *MemoryStorage) Set(key string, value string) (err error) {
	s.mutex.Lock()
	s.datamap[key] = value
	s.mutex.Unlock()
	return
}

func (s *MemoryStorage) MGet(keys ...string) (values []interface{}, err error) {
	values = make([]interface{}, len(keys))
	for i, key := range keys {
		values[i] = s.datamap[key]
	}
	return
}

func (s *MemoryStorage) MSet(keyvals ...string) (err error) {
	s.mutex.Lock()
	for i := 0; i < len(keyvals); i += 2 {
		key := keyvals[i]
		value := keyvals[i+1]
		s.datamap[key] = value
	}
	s.mutex.Unlock()
	return
}

// ============================================================
// 						Key
// ============================================================

func (s *MemoryStorage) Del(keys ...string) (n int, err error) {
	s.mutex.Lock()
	n = 0
	for _, key := range keys {
		_, exists := s.datamap[key]
		if exists {
			delete(s.datamap, key)
			n++
		}
	}
	s.mutex.Unlock()
	return
}

func (s *MemoryStorage) TypeOf(key string) (kt KeyType) {
	val, exist := s.datamap[key]
	if !exist {
		kt = KeyTypeUnknown
	} else {
		ty := val.(type)
		switch ty {
		case string:
			kt = KeyTypeString
		case *SafeList:
			kt = KeyTypeList
		default:
			kt = KeyTypeUnknown
		}
	}
	return
}

// ============================================================
// 						List
// ============================================================

// 获取指定key的列表，不存在时自动创建
func (s *MemoryStorage) listByKey(key string) (sl *SafeList) {
	s.mutex.Lock()
	var exists bool
	sl, exists = s.datamap[key]
	if !exists {
		sl = NewSafeList()
		s.datamap[key] = sl
	}
	s.mutex.Unlock()
	return
}

func (s *MemoryStorage) LPop(key string) (value interface{}, err error) {
	sl := s.listByKey(key)
	value = sl.LPop()
	return
}

func (s *MemoryStorage) LPush(key string, values ...string) (n int, err error) {
	sl := s.listByKey(key)
	n = sl.LPush(values...)
	return
}

func (s *MemoryStorage) RPop(key string) (value interface{}, err error) {
	sl := s.listByKey(key)
	value = sl.RPop()
	return
}

func (s *MemoryStorage) RPush(key string, values ...string) (n int, err error) {
	sl := s.listByKey(key)
	n = sl.RPush(values...)
	return
}

func (s *MemoryStorage) LRange(key string, start int, end int) (values []interface{}, err error) {
	sl := s.listByKey(key)
	values = sl.Range(start, end)
	return
}

func (s *MemoryStorage) LIndex(key string, index int) (value interface{}, err error) {
	sl := s.listByKey(key)
	value = sl.Index(index)
	return
}

func (s *MemoryStorage) LLen(key string) (length int, err error) {
	length = s.listByKey(key).Len()
	return
}

// ============================================================
// 						Hash
// ============================================================
// ============================================================
// 						Set
// ============================================================

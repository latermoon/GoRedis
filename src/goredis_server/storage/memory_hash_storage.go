package storage

import (
	"sync"
)

// 内存结构
type MapItem struct {
	Map   map[string]interface{}
	Mutex *sync.Mutex
}

func NewMapItem() (m *MapItem) {
	m = &MapItem{}
	m.Map = make(map[string]interface{})
	m.Mutex = &sync.Mutex{}
	return
}

// 暂时实现的map操作有同步问题，上线前需要修改
type MemoryHashStorage struct {
	HashStorage
	kvCache map[string]*MapItem
	mutex   *sync.Mutex
}

func NewMemoryHashStorage() (storage *MemoryHashStorage) {
	storage = &MemoryHashStorage{}
	storage.kvCache = make(map[string]*MapItem)
	storage.mutex = &sync.Mutex{}
	return
}

func (s *MemoryHashStorage) mapByKey(key string) (m *MapItem) {
	s.mutex.Lock()
	var exists bool
	m, exists = s.kvCache[key]
	if !exists {
		m = NewMapItem()
		s.kvCache[key] = m
	}
	s.mutex.Unlock()
	return
}

func (s *MemoryHashStorage) HGet(key string, field string) (value interface{}, err error) {
	m := s.mapByKey(key)
	m.Mutex.Lock()
	value, _ = m.Map[field]
	m.Mutex.Unlock()
	return
}

// http://redis.io/commands/hset
// Integer reply, specifically:
// 1 if field is a new field in the hash and value was set.
// 0 if field already exists in the hash and the value was updated.
func (s *MemoryHashStorage) HSet(key string, field string, value string) (result int, err error) {
	m := s.mapByKey(key)
	m.Mutex.Lock()
	_, exists := m.Map[field]
	m.Map[field] = value
	if !exists {
		result = 1
	} else {
		result = 0
	}
	m.Mutex.Unlock()
	return
}

// http://redis.io/commands/hgetall
func (s *MemoryHashStorage) HGetAll(key string) (keyvals []interface{}, err error) {
	m := s.mapByKey(key)
	m.Mutex.Lock()
	length := len(m.Map)
	keyvals = make([]interface{}, 0, length*2) // len(keys)+len(vals)
	for k, v := range m.Map {
		keyvals = append(keyvals, k)
		keyvals = append(keyvals, v)
	}
	m.Mutex.Unlock()
	return
}

func (s *MemoryHashStorage) HDel(key string, fields ...string) (n int, err error) {
	m := s.mapByKey(key)
	m.Mutex.Lock()
	n = 0
	for _, field := range fields {
		_, exists := m.Map[field]
		if exists {
			n++
			delete(m.Map, field)
		}
	}
	m.Mutex.Unlock()
	return
}

func (s *MemoryHashStorage) HMGet(key string, fields ...string) (values []interface{}, err error) {
	m := s.mapByKey(key)
	m.Mutex.Lock()
	values = make([]interface{}, len(fields))
	for i, field := range fields {
		values[i], _ = m.Map[field]
	}
	m.Mutex.Unlock()
	return
}

func (s *MemoryHashStorage) HMSet(key string, keyvals ...string) (err error) {
	m := s.mapByKey(key)
	m.Mutex.Lock()
	count := len(keyvals)
	for i := 0; i < count; i += 2 {
		k := keyvals[i]
		v := keyvals[i+1]
		m.Map[k] = v
	}
	m.Mutex.Unlock()
	return
}

func (s *MemoryHashStorage) HLen(key string) (length int, err error) {
	m := s.mapByKey(key)
	m.Mutex.Lock()
	length = len(m.Map)
	m.Mutex.Unlock()
	return
}

func (s *MemoryHashStorage) Del(keys ...string) (n int, err error) {
	s.mutex.Lock()
	n = 0
	for _, key := range keys {
		_, exists := s.kvCache[key]
		if exists {
			n++
			delete(s.kvCache, key)
		}
	}
	s.mutex.Unlock()
	return
}

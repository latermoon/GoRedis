package storage

import (
	"sync"
)

// 基于内存的StringStorage
type MemoryStringStorage struct {
	StringStorage
	kvCache map[string]interface{}
	mutex   *sync.Mutex
}

func NewMemoryStringStorage() (storage *MemoryStringStorage) {
	storage = &MemoryStringStorage{}
	storage.kvCache = make(map[string]interface{})
	storage.mutex = &sync.Mutex{}
	return
}

func (s *MemoryStringStorage) Get(key string) (value interface{}, err error) {
	value, _ = s.kvCache[key]
	return
}

func (s *MemoryStringStorage) Set(key string, value string) (err error) {
	s.mutex.Lock()
	s.kvCache[key] = value
	s.mutex.Unlock()
	return
}

func (s *MemoryStringStorage) MGet(keys ...string) (values []interface{}, err error) {
	values = make([]interface{}, len(keys))
	for i, key := range keys {
		values[i] = s.kvCache[key]
	}
	return
}

func (s *MemoryStringStorage) MSet(keyvals ...string) (err error) {
	s.mutex.Lock()
	for i := 0; i < len(keyvals); i += 2 {
		key := keyvals[i]
		value := keyvals[i+1]
		s.kvCache[key] = value
	}
	s.mutex.Unlock()
	return
}

func (s *MemoryStringStorage) Del(keys ...string) (n int, err error) {
	s.mutex.Lock()
	n = 0
	for _, key := range keys {
		_, exists := s.kvCache[key]
		if exists {
			delete(s.kvCache, key)
			n++
		}
	}
	s.mutex.Unlock()
	return
}

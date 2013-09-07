package storage

import (
	"errors"
	"sync"
)

type MemoryKeyTypeStorage struct {
	KeyTypeStorage
	caches map[string]KeyType
	mutex  *sync.Mutex
}

func NewMemoryKeyTypeStorage() (storage *MemoryKeyTypeStorage) {
	storage = &MemoryKeyTypeStorage{}
	storage.caches = make(map[string]KeyType)
	storage.mutex = &sync.Mutex{}
	return
}

func (s *MemoryKeyTypeStorage) GetType(key string) (keytype KeyType) {
	var exists bool
	keytype, exists = s.caches[key]
	if !exists {
		keytype = KeyTypeUnknown
	}
	return
}

func (s *MemoryKeyTypeStorage) SetType(key string, keytype KeyType) (err error) {
	s.mutex.Lock()
	tp := s.GetType(key)
	if tp == KeyTypeUnknown {
		s.caches[key] = keytype
	} else if tp != keytype {
		err = errors.New("Different from the former")
	}
	s.mutex.Unlock()
	return
}

func (s *MemoryKeyTypeStorage) DelType(key string) (keytype KeyType) {
	s.mutex.Lock()
	var exists bool
	keytype, exists = s.caches[key]
	if exists {
		delete(s.caches, key)
	} else {
		keytype = KeyTypeUnknown
	}
	s.mutex.Unlock()
	return
}

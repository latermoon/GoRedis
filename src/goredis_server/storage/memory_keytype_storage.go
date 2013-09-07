package storage

import (
	"errors"
)

type MemoryKeyTypeStorage struct {
	KeyTypeStorage
	caches   map[string]KeyType
	lockChan chan int
}

func NewMemoryKeyTypeStorage() (storage *MemoryKeyTypeStorage) {
	storage = &MemoryKeyTypeStorage{}
	storage.caches = make(map[string]KeyType)
	storage.lockChan = make(chan int, 1)
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
	s.lockChan <- 1
	tp := s.GetType(key)
	if tp == KeyTypeUnknown {
		s.caches[key] = keytype
	} else if tp != keytype {
		err = errors.New("Different from the former")
	}
	<-s.lockChan
	return
}

func (s *MemoryKeyTypeStorage) DelType(key string) (keytype KeyType) {
	s.lockChan <- 1
	var exists bool
	keytype, exists = s.caches[key]
	if exists {
		delete(s.caches, key)
	} else {
		keytype = KeyTypeUnknown
	}
	<-s.lockChan
	return
}

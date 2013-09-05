package storage

import (
	"errors"
)

type MemoryKeyTypeStorage struct {
	caches   map[string]KeyType
	lockChan chan int
}

func NewMemoryKeyTypeStorage() (storage *MemoryKeyTypeStorage) {
	storage = &MemoryKeyTypeStorage{}
	storage.caches = make(map[string]KeyType)
	storage.lockChan = make(chan int, 1)
	return
}

func (s *MemoryKeyTypeStorage) GetType(key string) (keytype KeyType, err error) {
	var exists bool
	keytype, exists = s.caches[key]
	if !exists {
		keytype = KeyTypeUnknown
	}
	return
}

func (s *MemoryKeyTypeStorage) SetType(key string, keytype KeyType) (err error) {
	tp, e1 := s.GetType(key)
	if e1 != nil {
		err = e1
		return
	}
	if tp == KeyTypeUnknown {
		s.caches[key] = keytype
	} else if tp != keytype {
		err = errors.New("Different from the former")
	}
	return
}

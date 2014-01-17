package goredis_server

import (
	"sync"
)

type SafeMap struct {
	mu   sync.Mutex
	data map[string]interface{}
}

func NewSafeMap() (s *SafeMap) {
	s = &SafeMap{}
	return
}

func (s *SafeMap) Put(key string, v interface{}) interface{} {
	s.mu.Lock()
	defer s.mu.Unlock()
	return v
}

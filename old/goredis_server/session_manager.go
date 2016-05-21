package goredis_server

import (
	"sync"
)

// 通用连接管理，本质是一个安全的map
type SessionManager struct {
	clients map[string]interface{}
	mu      sync.RWMutex
}

func NewSessionManager() (s *SessionManager) {
	s = &SessionManager{
		clients: map[string]interface{}{},
	}
	return
}

func (s *SessionManager) Put(key string, v interface{}) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.clients[key] = v
}

func (s *SessionManager) Get(key string) interface{} {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.clients[key]
}

func (s *SessionManager) Contains(key string) (ok bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	_, ok = s.clients[key]
	return
}

func (s *SessionManager) Remove(key string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	delete(s.clients, key)
}

func (s *SessionManager) Len() int {
	return len(s.clients)
}

func (s *SessionManager) Enumerate(fn func(i int, key string, val interface{})) {
	s.mu.RLock()
	// copy
	m := make(map[string]interface{})
	for k, v := range s.clients {
		m[k] = v
	}
	s.mu.RUnlock()

	i := 0
	for k, v := range m {
		fn(i, k, v)
		i++
	}
}

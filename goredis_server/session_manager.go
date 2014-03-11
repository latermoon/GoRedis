package goredis_server

import (
	. "GoRedis/goredis"
	"sync"
)

// 管理当前连入的客户端
type SessionManager struct {
	clients map[string]*Session
	mu      sync.RWMutex
}

func NewSessionManager() (s *SessionManager) {
	s = &SessionManager{
		clients: map[string]*Session{},
	}
	return
}

func (s *SessionManager) Put(key string, session *Session) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.clients[key] = session
}

func (s *SessionManager) Get(key string) *Session {
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

func (s *SessionManager) Enumerate(fn func(i int, key string, session *Session)) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	i := 0
	for k, v := range s.clients {
		fn(i, k, v)
		i++
	}
}

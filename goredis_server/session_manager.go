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

func (s *SessionManager) Put(host string, session *Session) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.clients[host] = session
}

func (s *SessionManager) Contains(host string) (ok bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	_, ok = s.clients[host]
	return
}

func (s *SessionManager) Remove(host string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	delete(s.clients, host)
}

func (s *SessionManager) Len() int {
	return len(s.clients)
}

func (s *SessionManager) Enumerate(fn func(i int, host string, session *Session)) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	i := 0
	for k, v := range s.clients {
		fn(i, k, v)
		i++
	}
}

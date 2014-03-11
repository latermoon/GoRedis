package goredis_server

// 管理多个从库SyncClient对象
import (
	. "GoRedis/goredis"
	"sync"
)

type SyncManager struct {
	clients map[string]*Session
	mu      sync.RWMutex
}

func NewSyncManager() (s *SyncManager) {
	s = &SyncManager{
		clients: map[string]*Session{},
	}
	return
}

func (s *SyncManager) Put(host string, sess *Session) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.clients[host] = sess
}

func (s *SyncManager) Contains(host string) (ok bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	_, ok = s.clients[host]
	return
}

func (s *SyncManager) Enumerate(fn func(i int, host string, sess *Session)) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	i := 0
	for k, v := range s.clients {
		fn(i, k, v)
		i++
	}
}

func (s *SyncManager) Remove(host string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	delete(s.clients, host)
}

func (s *SyncManager) Len() int {
	return len(s.clients)
}

package goredis_server

import (
	// . "GoRedis/goredis"
	"container/list"
	"sync"
	"time"
)

type SlaveManager struct {
	clients *list.List
	mu      sync.Mutex
}

func NewSlaveManager() (s *SlaveManager) {
	s = &SlaveManager{
		clients: list.New(),
	}
	go s.pingRunloop()
	return
}

func (s *SlaveManager) pingRunloop() {
	ticker := time.NewTicker(time.Second * 1)
	go func() {
		for _ = range ticker.C {

		}
	}()
}

func (s *SlaveManager) Count() int {
	return s.clients.Len()
}

func (s *SlaveManager) Client(i int) (c *SlaveClient) {
	s.mu.Lock()
	defer s.mu.Unlock()
	cur := 0
	for e := s.clients.Front(); e != nil; e = e.Next() {
		c := e.Value.(*SlaveClient)
		if i == cur {
			return c
		}
		cur++
	}
	return
}

package goredis_server

import (
	// . "GoRedis/goredis"
	"GoRedis/libs/stdlog"
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
	go s.checkRunloop()
	return
}

func (s *SlaveManager) checkRunloop() {
	ticker := time.NewTicker(time.Second * 1)
	for _ = range ticker.C {
		s.mu.Lock()
		for e := s.clients.Front(); e != nil; e = e.Next() {
			c := e.Value.(*SlaveClient)
			if c.Broken() {
				stdlog.Printf("[M %s] master broken, removed\n", c.session.RemoteAddr())
				s.clients.Remove(e)
			}
		}
		s.mu.Unlock()
	}
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

func (s *SlaveManager) Add(c *SlaveClient) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.clients.PushBack(c)
}

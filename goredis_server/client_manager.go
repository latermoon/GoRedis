package goredis_server

import (
	. "GoRedis/goredis"
	"sync"
)

// 管理当前连入的客户端
type ClientManager struct {
	clients map[string]*Session
	mu      sync.RWMutex
}

func NewClientManager() (c *ClientManager) {
	c = &ClientManager{
		clients: map[string]*Session{},
	}
	return
}

func (c *ClientManager) Put(host string, session *Session) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.clients[host] = session
}

func (c *ClientManager) Remove(host string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	delete(c.clients, host)
}

func (c *ClientManager) Len() int {
	return len(c.clients)
}

func (c *ClientManager) Enumerate(fn func(i int, host string, session *Session)) {
	c.mu.RLock()
	defer c.mu.RUnlock()
	i := 0
	for k, v := range c.clients {
		fn(i, k, v)
		i++
	}
}

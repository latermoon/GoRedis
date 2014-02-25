package goredis_server

import (
	. "GoRedis/goredis"
	"GoRedis/libs/stdlog"
	"errors"
	"sync"
)

type MonClient struct {
	session *Session
	buffer  chan string
	broken  bool
	mu      sync.Mutex
}

func NewMonClient(session *Session) (m *MonClient) {
	m = &MonClient{
		session: session,
		buffer:  make(chan string, 10000),
		broken:  false,
	}
	go m.runloop()
	return
}

func (m *MonClient) Broken() bool {
	return m.broken
}

func (m *MonClient) Send(line string) (err error) {
	if m.broken {
		return errors.New("monitor broken")
	}
	if len(m.buffer) == cap(m.buffer) {
		m.destory()
		return errors.New("out of buffer limit")
	}
	m.mu.Lock()
	m.buffer <- line
	m.mu.Unlock()
	return
}

func (m *MonClient) runloop() {
	for {
		if m.broken {
			break
		}
		line, ok := <-m.buffer
		if !ok {
			break
		}
		err := m.session.WriteReply(StatusReply(line))
		if err != nil {
			m.destory()
			break
		}
	}
	stdlog.Printf("[%s] monitor runloop broken", m.session.RemoteAddr())
}

// 释放资源
func (m *MonClient) destory() {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.broken = true
	close(m.buffer)
	// clear up
	for {
		_, ok := <-m.buffer
		if !ok {
			break
		}
	}
}

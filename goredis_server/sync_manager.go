package goredis_server

// 管理多个从库SyncClient对象
import (
	. "GoRedis/goredis"
	"container/list"
	"sync"
	"time"
)

type SyncManager struct {
	clients *list.List
	mu      sync.Mutex
}

func NewSyncManager() (s *SyncManager) {
	s = &SyncManager{
		clients: list.New(),
	}
	go s.pingRunloop()
	return
}

func (s *SyncManager) pingRunloop() {
	ticker := time.NewTicker(time.Second * 1)
	go func() {
		for _ = range ticker.C {
			s.BroadcastCommand(NewCommand([]byte("PING")))
		}
	}()
}

func (s *SyncManager) Count() int {
	return s.clients.Len()
}

// 广播同步
func (s *SyncManager) BroadcastCommand(cmd *Command) (n int) {
	if s.clients.Len() == 0 {
		return
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	for e := s.clients.Front(); e != nil; e = e.Next() {
		c := e.Value.(*SyncClient)
		err := c.Enqueue(cmd)
		if err != nil {
			errlog.Println("[S %s] broadcast error %s", c.session.RemoteAddr(), err)
			s.clients.Remove(e)
		} else {
			n++
		}
	}
	return
}

func (s *SyncManager) Client(i int) (c *SyncClient) {
	s.mu.Lock()
	defer s.mu.Unlock()
	cur := 0
	for e := s.clients.Front(); e != nil; e = e.Next() {
		c := e.Value.(*SyncClient)
		if i == cur {
			return c
		}
		cur++
	}
	return
}

func (s *SyncManager) Add(c *SyncClient) {
	s.mu.Lock()
	defer s.mu.Unlock()
	_, exist := s.exist(c)
	if exist {
		return
	}
	s.clients.PushBack(c)
}

func (s *SyncManager) Remove(c *SyncClient) {
	s.mu.Lock()
	defer s.mu.Unlock()
	e, exist := s.exist(c)
	if exist {
		s.clients.Remove(e)
	}
}

func (s *SyncManager) Exist(c *SyncClient) (*list.Element, bool) {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.exist(c)
}

func (s *SyncManager) exist(c *SyncClient) (*list.Element, bool) {
	for e := s.clients.Front(); e != nil; e = e.Next() {
		if e.Value.(*SyncClient) == c {
			return e, true
		}
	}
	return nil, false
}

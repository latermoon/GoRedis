package goredis_server

// 管理SlaveClient对象
import (
	. "GoRedis/goredis"
	"GoRedis/libs/stdlog"
	"container/list"
	"errors"
	"net"
	"strings"
	"sync"
	"time"
)

type ISlaveClient interface {
	Sync(uid string) (err error)
	Broken() bool
	RemoteAddr() net.Addr
	Status() string
}

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

// 判断是否已存在目标连接
func (s *SlaveManager) Contains(host string) bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	for e := s.clients.Front(); e != nil; e = e.Next() {
		c := e.Value.(ISlaveClient)
		if host == c.RemoteAddr().String() {
			return true
		}
	}
	return false
}

func (s *SlaveManager) checkRunloop() {
	ticker := time.NewTicker(time.Second * 1)
	for _ = range ticker.C {
		s.mu.Lock()
		for e := s.clients.Front(); e != nil; e = e.Next() {
			c := e.Value.(ISlaveClient)
			if c.Broken() {
				stdlog.Printf("[M %s] master broken, removed\n", c.RemoteAddr())
				s.clients.Remove(e)
			}
		}
		s.mu.Unlock()
	}
}

func (s *SlaveManager) Count() int {
	return s.clients.Len()
}

func (s *SlaveManager) Client(i int) (c ISlaveClient) {
	s.mu.Lock()
	defer s.mu.Unlock()
	cur := 0
	for e := s.clients.Front(); e != nil; e = e.Next() {
		c := e.Value.(ISlaveClient)
		if i == cur {
			return c
		}
		cur++
	}
	return
}

func (s *SlaveManager) Add(c ISlaveClient) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.clients.PushBack(c)
}

func redisInfo(session *Session) (isgoredis bool, version string, err error) {
	cmdinfo := NewCommand([]byte("info"), []byte("server"))
	session.WriteCommand(cmdinfo)
	var reply *Reply
	reply, err = session.ReadReply()
	if err != nil {
		return
	}
	if reply.Value == nil {
		err = errors.New("reply nil")
		return
	}

	var info string
	switch reply.Value.(type) {
	case string:
		info = reply.Value.(string)
	case []byte:
		info = string(reply.Value.([]byte))
	default:
		info = reply.String()
	}

	// 切分info返回的数据，存放到map里
	kv := make(map[string]string)
	lines := strings.Split(info, "\n")
	for _, line := range lines {
		line = strings.TrimSuffix(line, "\r")
		line = strings.TrimPrefix(line, " ")
		if strings.HasPrefix(line, "#") {
			continue
		}
		pairs := strings.Split(line, ":")
		if len(pairs) != 2 {
			continue
		}
		// done
		kv[pairs[0]] = pairs[1]
	}

	_, isgoredis = kv["goredis_version"]
	if isgoredis {
		version = kv["goredis_version"]
	} else {
		version = kv["redis_version"]
	}

	return
}

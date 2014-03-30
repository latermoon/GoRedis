package goredis_proxy

import (
	. "GoRedis/goredis"
	"GoRedis/libs/counter"
	"crypto/md5"
	"net"
	"sync"
	"time"
)

type RemoteInfo struct {
	Ops_per_sec    int64
	last_total_ops int64
	Uptime         time.Time
}

// RemoteSession表示一个远程会话
type RemoteSession struct {
	host      string
	poolSize  int
	mus       []*sync.Mutex
	sessions  []*Session
	counters  *counter.Counters
	available bool
	Info      *RemoteInfo
	ticker    *time.Ticker
}

func NewRemoteSession(host string, poolSize int) (s *RemoteSession, err error) {
	s = &RemoteSession{
		host:      host,
		poolSize:  poolSize,
		counters:  counter.NewCounters(),
		available: true,
		Info: &RemoteInfo{
			Uptime: time.Now(),
		},
	}
	s.mus = make([]*sync.Mutex, s.poolSize)
	s.sessions = make([]*Session, s.poolSize)
	for i := 0; i < s.poolSize; i++ {
		s.mus[i] = &sync.Mutex{}
		s.sessions[i], err = s.createSession()
		if err != nil {
			s.available = false
			break
		}
	}
	go s.secondTicker()
	return
}

func (s *RemoteSession) secondTicker() {
	s.ticker = time.NewTicker(time.Second * 1)
	for _ = range s.ticker.C {
		// ops_per_sec
		total := s.counters.Get("total").Count()
		s.Info.Ops_per_sec, s.Info.last_total_ops = total-s.Info.last_total_ops, total
	}
}

// 发送指令到远程Redis，并返回结果
func (s *RemoteSession) Invoke(session *Session, cmd *Command) (reply *Reply, err error) {
	i := s.indexOf([]byte(session.RemoteAddr().String()))
	s.counters.Get("total").Incr(1)
	// lock
	s.mus[i].Lock()
	defer s.mus[i].Unlock()
	// redirect
	err = s.sessions[i].WriteCommand(cmd)
	if err == nil {
		reply, err = s.sessions[i].ReadReply()
	}
	if err != nil {
		s.available = false
	}
	return
}

func (s *RemoteSession) Available() bool {
	return s.available
}

func (s *RemoteSession) RemoteAddr() string {
	return s.sessions[0].RemoteAddr().String()
}

func (s *RemoteSession) Close() {
	s.available = false
	for _, session := range s.sessions {
		session.Close()
	}
	s.ticker.Stop()
}

func (s *RemoteSession) createSession() (session *Session, err error) {
	var conn net.Conn
	conn, err = net.Dial("tcp", s.host)
	if err != nil {
		return
	}
	session = NewSession(conn)
	return
}

func (s *RemoteSession) indexOf(key []byte) (i int) {
	hash := md5.Sum(key)
	sum := 0
	for i, count := 0, len(hash); i < count; i++ {
		sum += int(hash[i]) << uint(i)
	}
	i = sum % len(s.mus)
	return
}

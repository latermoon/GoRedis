package goredis_proxy

import (
	. "GoRedis/goredis"
	"GoRedis/libs/counter"
	"crypto/md5"
	"errors"
	"net"
	"sync"
	"time"
)

type RemoteInfo struct {
	Ops_per_sec        int64
	TotalCommands      int64
	last_total_ops     int64
	Uptime             time.Time
	LastCommandIsWrite bool
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
	if s.available {
		go s.secondTicker()
	}
	return
}

func (s *RemoteSession) secondTicker() {
	s.ticker = time.NewTicker(time.Second * 1)
	for _ = range s.ticker.C {
		// ops_per_sec
		s.Info.TotalCommands = s.counters.Get("total").Count()
		s.Info.Ops_per_sec, s.Info.last_total_ops = s.Info.TotalCommands-s.Info.last_total_ops, s.Info.TotalCommands
	}
}

// 发送指令到远程Redis，并返回结果
func (s *RemoteSession) Send(session *Session, cmd *Command) (reply *Reply, err error) {
	if !s.available {
		err = errors.New("unavailable")
		return
	}
	i := s.indexOf([]byte(session.RemoteAddr().String()))
	// lock
	s.mus[i].Lock()
	defer s.mus[i].Unlock()

	s.counters.Get("total").Incr(1)
	// redirect
	err = s.sessions[i].WriteCommand(cmd)
	if err == nil {
		reply, err = s.sessions[i].ReadReply()
	}
	if err != nil {
		s.counters.Get("error").Incr(1)
		s.sessions[i].Close()
		s.sessions[i], err = s.createSession() // reconnect
		if err != nil {
			s.available = false
		}
	}
	return
}

func (s *RemoteSession) Available() bool {
	return s.available
}

func (s *RemoteSession) RemoteAddr() string {
	if !s.available {
		return s.host
	}
	return s.sessions[0].RemoteAddr().String()
}

func (s *RemoteSession) Close() {
	if !s.available {
		return
	}
	s.available = false
	s.Info.LastCommandIsWrite = false
	s.Info.Ops_per_sec = 0
	for _, session := range s.sessions {
		session.Close()
	}
	s.ticker.Stop()
}

func (s *RemoteSession) createSession() (session *Session, err error) {
	var conn net.Conn
	conn, err = net.DialTimeout("tcp", s.host, time.Millisecond*1000)
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

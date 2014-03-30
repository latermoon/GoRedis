package goredis_proxy

import (
	. "GoRedis/goredis"
	// "GoRedis/libs/stdlog"
	"GoRedis/libs/counter"
	"crypto/md5"
	"fmt"
	"net"
	"strconv"
	"sync"
)

// RemoteSession表示一个远程会话
type RemoteSession struct {
	host     string
	maxIdle  int
	mus      []*sync.Mutex
	sessions []*Session
	counters *counter.Counters
}

func NewRemoteSession(host string) (s *RemoteSession, err error) {
	s = &RemoteSession{
		host:     host,
		maxIdle:  100,
		counters: counter.NewCounters(),
	}
	s.mus = make([]*sync.Mutex, s.maxIdle)
	s.sessions = make([]*Session, s.maxIdle)
	for i := 0; i < s.maxIdle; i++ {
		s.mus[i] = &sync.Mutex{}
		s.sessions[i], err = s.newSession()
		if err != nil {
			break
		}
	}
	return
}

func (s *RemoteSession) Invoke(session *Session, cmd *Command) (reply *Reply, err error) {
	i := s.indexOf([]byte(session.RemoteAddr().String()))
	s.counters.Get(strconv.Itoa(i)).Incr(1)
	// lock
	s.mus[i].Lock()
	defer s.mus[i].Unlock()
	// redirect
	err = s.sessions[i].WriteCommand(cmd)
	if err == nil {
		reply, err = s.sessions[i].ReadReply()
	}
	return
}

func (s *RemoteSession) newSession() (session *Session, err error) {
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

func (s *RemoteSession) LockInfo() (lines []string) {
	lines = make([]string, s.maxIdle)
	for i := 0; i < s.maxIdle; i++ {
		lines[i] = fmt.Sprintf("%d=%d", i, s.counters.Get(strconv.Itoa(i)).Count())
	}
	return
}

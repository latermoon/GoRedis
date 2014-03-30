package goredis_proxy

import (
	. "GoRedis/goredis"
	"GoRedis/goredis_server"
	"GoRedis/libs/counter"
	"GoRedis/libs/stdlog"
	"runtime/debug"
	"sync"
)

const VERSION = "1.0.2"

// Redis代理
type GoRedisProxy struct {
	ServerHandler
	RedisServer
	// option
	options *Options
	// counter
	counters *counter.Counters
	// mgr
	sessmgr *goredis_server.SessionManager
	// sessions
	master *RemoteSession
	slave  *RemoteSession
	// lock
	rwlock sync.Mutex
}

func NewProxy(opt *Options) (s *GoRedisProxy) {
	s = &GoRedisProxy{}
	s.SetHandler(s)
	s.options = opt
	s.counters = counter.NewCounters()
	s.sessmgr = goredis_server.NewSessionManager()
	return
}

func (server *GoRedisProxy) Listen(host string) error {
	return server.RedisServer.Listen(host)
}

// ServerHandler.SessionOpened()
func (server *GoRedisProxy) SessionOpened(session *Session) {
	server.counters.Get("connection").Incr(1)
	server.sessmgr.Put(session.RemoteAddr().String(), session)
	stdlog.Println("connection accepted from", session.RemoteAddr())
}

// ServerHandler.SessionClosed()
func (server *GoRedisProxy) SessionClosed(session *Session, err error) {
	server.counters.Get("connection").Incr(-1)
	server.sessmgr.Remove(session.RemoteAddr().String())
	stdlog.Println("end connection", session.RemoteAddr(), err)
}

// ServerHandler.ExceptionCaught()
func (server *GoRedisProxy) ExceptionCaught(err error) {
	stdlog.Printf("exception %s\n", err)
	stdlog.Println(string(debug.Stack()))
}

// ServerHandler.On()
func (server *GoRedisProxy) On(session *Session, cmd *Command) (reply *Reply) {
	// for Suspend & Resume
	server.rwlock.Lock()
	server.rwlock.Unlock()

	cmdName := cmd.Name()
	switch cmdName {
	case "CONFIG":
		return server.OnCONFIG(session, cmd)
	case "INFO":
		return server.OnINFO(session, cmd)
	}

	var err error
	var remote *RemoteSession

	// dispatch
	if goredis_server.NeedSync(cmdName) {
		remote = server.master
	} else {
		remote = server.slave
	}

	// process
	reply, err = remote.Invoke(session, cmd)
	if err != nil {
		reply = ErrorReply(err)
	}
	return
}

// 挂起
func (server *GoRedisProxy) Suspend() {
	server.rwlock.Lock()
}

// 恢复
func (server *GoRedisProxy) Resume() {
	server.rwlock.Unlock()
}

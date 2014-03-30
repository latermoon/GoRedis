package goredis_proxy

import (
	. "GoRedis/goredis"
	"GoRedis/goredis_server"
	"GoRedis/libs/counter"
	"GoRedis/libs/stdlog"
	"runtime/debug"
)

const VERSION = "1.0.1"

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
	cmdName := cmd.Name()
	if cmdName == "INFO" {
		return server.OnINFO(session, cmd)
	}

	var err error
	var remote *RemoteSession

	if goredis_server.NeedSync(cmdName) {
		remote = server.master
	} else {
		remote = server.slave
	}

	reply, err = remote.Invoke(session, cmd)
	if err != nil {
		stdlog.Println("invoke error", err)
		reply = ErrorReply(err)
	}
	return
}

func (server *GoRedisProxy) OnINFO(session *Session, cmd *Command) (reply *Reply) {
	lines := server.master.LockInfo()
	bulks := make([]interface{}, len(lines))
	for i, line := range lines {
		bulks[i] = line
	}
	reply = MultiBulksReply(bulks)
	return
}

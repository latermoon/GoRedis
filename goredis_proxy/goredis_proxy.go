package goredis_proxy

import (
	. "GoRedis/goredis"
	"GoRedis/goredis_server"
	"GoRedis/libs/counter"
	"GoRedis/libs/stdlog"
	"errors"
	"math/rand"
	"runtime/debug"
	"sync"
)

const VERSION = "1.0.4"

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

	// preprocess
	reply = server.invokeCommand(session, cmd)
	if reply != nil {
		return
	}

	if cmd.Len() < 2 {
		return ErrorReply("not support")
	}

	cmdName := cmd.Name()
	key := cmd.StringAtIndex(1)

	// dispatch
	var err error
	if isWriteAction(cmdName) {
		// 写入主库
		if server.master.Available() {
			if server.options.CanWrite() {
				reply, err = server.master.Invoke(session, cmd)
				session.SetAttribute(S_LAST_WRITE_KEY, key)
			} else {
				err = errors.New("reject write")
			}
		} else {
			err = errors.New("master gone away")
		}
	} else {
		// 默认顺序先从库，再主库
		remotes := []*RemoteSession{server.slave, server.master}
		lastWriteKey, ok := session.GetAttribute(S_LAST_WRITE_KEY).(string)
		// 如果上次是写操作，然后再读取相同的key，优先访问主库，保障一致性
		// 如果可以读主库，则随机读取，否则读从库
		// 如果第一选择访问失败，则尝试第二选择而不论主从
		if (ok && lastWriteKey == key) || (server.options.CanReadMaster() && rand.Intn(2) == 1) {
			remotes[0], remotes[1] = remotes[1], remotes[0]
		}
		for _, remote := range remotes {
			reply, err = remote.Invoke(session, cmd)
			session.SetAttribute(S_LAST_WRITE_KEY, "")
			if err == nil {
				break
			}
		}
	}

	// check
	if err != nil {
		reply = ErrorReply(err)
	}
	return
}

func (server *GoRedisProxy) invokeCommand(session *Session, cmd *Command) (reply *Reply) {
	cmdName := cmd.Name()
	switch cmdName {
	case "CONFIG":
		return server.OnCONFIG(session, cmd)
	case "INFO":
		return server.OnINFO(session, cmd)
	case "PING":
		return StatusReply("PONG")
	}
	if ignoreSync[cmdName] {
		return ErrorReply("not support")
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

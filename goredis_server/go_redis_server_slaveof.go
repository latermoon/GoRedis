package goredis_server

import (
	. "GoRedis/goredis"
	"GoRedis/libs/stdlog"
	"fmt"
	"net"
	"runtime/debug"
	"strings"
)

// 从主库获取数据
func (server *GoRedisServer) OnSLAVEOF(session *Session, cmd Command) (reply Reply) {
	// 保障不会奔溃
	defer func() {
		if v := recover(); v != nil {
			stdlog.Printf("[%s] slaveof panic %s\n", session.RemoteAddr(), cmd)
			stdlog.Println(string(debug.Stack()))
		}
	}()
	arg1, arg2 := cmd.StringAtIndex(1), cmd.StringAtIndex(2)
	// SLAVEOF NO ONE
	if strings.ToUpper(arg1) == "NO" && strings.ToUpper(arg2) == "ONE" {
		return server.onSlaveOfNoOne(session, cmd)
	}

	// connect to master
	hostPort := arg1 + ":" + arg2
	conn, err := net.Dial("tcp", hostPort)
	if err != nil {
		return ErrorReply(err)
	}

	// check exists
	remoteHost := conn.RemoteAddr().String()
	if server.slavemgr.Contains(remoteHost) {
		return ErrorReply("connection exists")
	}

	masterSession := NewSession(conn)
	isgoredis, version, err := redisInfo(masterSession)
	if err != nil {
		return ErrorReply(err)
	}

	var client ISlaveClient
	if isgoredis {
		slavelog.Printf("[M %s] SLAVEOF %s GoRedis:%s\n", remoteHost, remoteHost, version)
		if client, err = NewSlaveClientV2(server, masterSession); err != nil {
			return ErrorReply(err)
		}
	} else {
		slavelog.Printf("[M %s] SLAVEOF %s Redis:%s\n", remoteHost, remoteHost, version)
		if client, err = NewSlaveClient(server, masterSession); err != nil {
			return ErrorReply(err)
		}
	}

	// async
	go func() {
		client.Session().SetAttribute(S_STATUS, REPL_WAIT)
		server.slavemgr.Put(remoteHost, client)
		err := client.Sync()
		if err != nil {
			slavelog.Printf("[M %s] sync broken %s\n", remoteHost, err)
		}
		client.Close()
		server.slavemgr.Remove(remoteHost)
	}()

	return StatusReply("OK")
}

// SLAVEOF NO ONE will stop replication
func (server *GoRedisServer) onSlaveOfNoOne(session *Session, cmd Command) (reply Reply) {
	slavelog.Printf("SLAVEOF NO ONE, will disconnect %d connection(s)\n", server.slavemgr.Len())
	reply = StatusReply(fmt.Sprintf("disconnect %d connections(s)", server.slavemgr.Len()))

	server.slavemgr.Enumerate(func(i int, key string, val interface{}) {
		client := val.(ISlaveClient)
		client.Close()
		server.slavemgr.Remove(key)
	})
	return
}

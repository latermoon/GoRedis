package goredis_server

import (
	. "GoRedis/goredis"
	"net"
)

// 从主库获取数据
// 对应 go_redis_server_sync.go
func (server *GoRedisServer) OnSLAVEOF(session *Session, cmd *Command) (reply *Reply) {
	// connect to master
	host := cmd.StringAtIndex(1)
	port := cmd.StringAtIndex(2)
	hostPort := host + ":" + port

	conn, err := net.Dial("tcp", hostPort)
	if err != nil {
		return ErrorReply(err)
	}

	// 异步处理
	masterSession := NewSession(conn)
	isgoredis, version, err := redisInfo(masterSession)
	if err != nil {
		return ErrorReply(err)
	}

	var client ISlaveClient
	if isgoredis {
		slavelog.Printf("[M %s] SLAVEOF %s GoRedis:%s\n", masterSession.RemoteAddr(), masterSession.RemoteAddr(), version)
		if client, err = NewSlaveClientV2(server, masterSession); err != nil {
			return ErrorReply(err)
		}
	} else {
		slavelog.Printf("[M %s] SLAVEOF %s Redis:%s\n", masterSession.RemoteAddr(), masterSession.RemoteAddr(), version)
		if client, err = NewSlaveClient(server, masterSession); err != nil {
			return ErrorReply(err)
		}
	}
	server.slavemgr.Add(client)
	go client.Sync(server.UID())

	return StatusReply("OK")
}

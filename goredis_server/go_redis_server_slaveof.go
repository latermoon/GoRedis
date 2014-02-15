package goredis_server

import (
	. "GoRedis/goredis"
	"GoRedis/libs/stdlog"
	"net"
)

// 从主库获取数据
// 对应 go_redis_server_sync.go
func (server *GoRedisServer) OnSLAVEOF(cmd *Command) (reply *Reply) {
	// connect to master
	host := cmd.StringAtIndex(1)
	port := cmd.StringAtIndex(2)
	hostPort := host + ":" + port

	stdlog.Println("SLAVEOF:", hostPort)

	conn, err := net.Dial("tcp", hostPort)
	if err != nil {
		reply = ErrorReply(err)
		stdlog.Println(err)
		return
	}
	reply = StatusReply("OK")
	// 异步处理
	slaveClient, err := NewSlaveClient(server, NewSession(conn))
	if err != nil {
		reply = ErrorReply(err)
		return
	}
	go slaveClient.Sync(server.UID())
	return
}

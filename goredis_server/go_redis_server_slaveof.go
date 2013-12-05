package goredis_server

import (
	. "../goredis"
	"net"
)

// 从主库获取数据
// 对应 go_redis_server_sync.go
func (server *GoRedisServer) OnSLAVEOF(cmd *Command) (reply *Reply) {
	// connect to master
	host := cmd.StringAtIndex(1)
	port := cmd.StringAtIndex(2)
	hostPort := host + ":" + port

	conn, err := net.Dial("tcp", hostPort)
	if err != nil {
		reply = ErrorReply(err)
		return
	}
	reply = StatusReply("OK")
	// 异步处理
	slaveSession := NewSlaveSession(NewSession(conn), hostPort)
	slaveClient := NewSlaveSessionClient(server, slaveSession)
	go slaveClient.Start()
	return
}

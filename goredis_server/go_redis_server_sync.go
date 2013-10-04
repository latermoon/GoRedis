package goredis_server

import (
	. "../goredis"
	// . "./storage"
	"fmt"
)

// 向从库发送数据
// 对应 go_redis_server_slaveof.go
func (server *GoRedisServer) OnSYNC(cmd *Command, session *Session) (reply *Reply) {
	fmt.Println("recv sync")

	return StatusReply("OK")
}

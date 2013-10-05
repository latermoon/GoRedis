package goredis_server

import (
	. "../goredis"
	. "./storage"
	"fmt"
	"strings"
)

// 向从库发送数据
// SYNC uid 70ecc21580
// 对应 go_redis_server_slaveof.go
func (server *GoRedisServer) OnSYNC(cmd *Command, session *Session) (reply *Reply) {
	// 客户端标识 SYNC uid 70ecc21580
	uid := ""
	args := cmd.StringArgs()
	if len(args) >= 3 && strings.ToLower(args[1]) == "uid" {
		uid = args[2]
	}
	fmt.Println("new slave ", uid)

	snapshot, err := server.datasource.(*LevelDBDataSource).DB().GetSnapshot()
	if err != nil {
		return ErrorReply(err)
	}

	slave := NewSlaveSession(server, session, uid)
	server.slavelist.PushBack(slave)

	go slave.SendSnapshot(snapshot)

	// SYNC不需要Reply
	reply = nil
	return
}

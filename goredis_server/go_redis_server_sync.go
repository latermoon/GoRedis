package goredis_server

import (
	. "../goredis"
	. "./storage"
	"fmt"
)

// 向从库发送数据
// 对应 go_redis_server_slaveof.go
func (server *GoRedisServer) OnSYNC(cmd *Command, session *Session) (reply *Reply) {
	fmt.Println("recv sync ...")

	snapshot, err := server.datasource.(*LevelDBDataSource).DB().GetSnapshot()
	if err != nil {
		return ErrorReply(err)
	}

	slave := NewSlaveSession(session)
	server.slavelist.PushBack(slave)

	go slave.SendSnapshot(snapshot)

	// SYNC不需要Reply
	reply = nil
	return
}

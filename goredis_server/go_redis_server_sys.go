package goredis_server

import (
	. "../goredis"
)

func (server *GoRedisServer) OnPING(cmd *Command) (reply *Reply) {
	reply = StatusReply("PONG")
	return
}

func (server *GoRedisServer) OnINFO(cmd *Command) (reply *Reply) {
	reply = BulkReply("GoRedis by latermoon")
	return
}

func (server *GoRedisServer) OnAUTH(cmd *Command) (reply *Reply) {
	password := cmd.StringAtIndex(1)
	if password == "GoRedis" {
		reply = StatusReply("OK")
	} else {
		reply = ErrorReply("403")
	}
	return
}

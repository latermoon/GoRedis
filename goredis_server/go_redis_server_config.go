package goredis_server

import (
	. "github.com/latermoon/GoRedis/goredis"
	"strings"
)

// http://redis.io/commands#server
func (server *GoRedisServer) OnCONFIG(cmd *Command) (reply *Reply) {
	action := strings.ToUpper(cmd.StringAtIndex(1))
	switch action {
	case "GET":
		reply = server.configGet(cmd)
	case "SET":
		reply = server.configSet(cmd)
	case "REWRITE":
		reply = server.configRewrite(cmd)
	case "RESETSTAT":
		reply = server.configResetStat(cmd)
	default:
		reply = ErrorReply("Bad Action")
	}
	return
}

func (server *GoRedisServer) configGet(cmd *Command) (reply *Reply) {
	name := cmd.StringAtIndex(2)
	reply = BulkReply(name)
	return
}

func (server *GoRedisServer) configSet(cmd *Command) (reply *Reply) {
	return
}

func (server *GoRedisServer) configRewrite(cmd *Command) (reply *Reply) {
	return
}

func (server *GoRedisServer) configResetStat(cmd *Command) (reply *Reply) {
	return
}

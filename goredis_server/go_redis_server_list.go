package goredis_server

import (
	. "../goredis"
	"strconv"
)

func (server *GoRedisServer) OnLLEN(cmd *Command) (reply *Reply) {
	key := cmd.StringAtIndex(1)
	length, err := server.Storage.LLen(key)
	reply = ReplySwitch(err, IntegerReply(length))
	return
}

func (server *GoRedisServer) OnLINDEX(cmd *Command) (reply *Reply) {
	key := cmd.StringAtIndex(1)
	index, _ := strconv.Atoi(cmd.StringAtIndex(2))
	value, err := server.Storage.LIndex(key, index)
	reply = ReplySwitch(err, BulkReply(value))
	return
}

func (server *GoRedisServer) OnLRANGE(cmd *Command) (reply *Reply) {
	key := cmd.StringAtIndex(1)
	start, _ := strconv.Atoi(cmd.StringAtIndex(2))
	end, _ := strconv.Atoi(cmd.StringAtIndex(3))
	values, err := server.Storage.LRange(key, start, end)
	reply = ReplySwitch(err, MultiBulksReply(values))
	return
}

func (server *GoRedisServer) OnRPUSH(cmd *Command) (reply *Reply) {
	key := cmd.StringAtIndex(1)
	values := cmd.StringArgs()[2:]
	n, err := server.Storage.RPush(key, values...)
	reply = ReplySwitch(err, IntegerReply(n))
	return
}

func (server *GoRedisServer) OnLPOP(cmd *Command) (reply *Reply) {
	key := cmd.StringAtIndex(1)
	value, err := server.Storage.LPop(key)
	if value == nil {
		server.Storage.Del(key)
	}
	reply = ReplySwitch(err, BulkReply(value))
	return
}

func (server *GoRedisServer) OnLPUSH(cmd *Command) (reply *Reply) {
	key := cmd.StringAtIndex(1)
	values := cmd.StringArgs()[2:]
	n, err := server.Storage.LPush(key, values...)
	reply = ReplySwitch(err, IntegerReply(n))
	return
}

func (server *GoRedisServer) OnRPOP(cmd *Command) (reply *Reply) {
	key := cmd.StringAtIndex(1)
	value, err := server.Storage.RPop(key)
	if value == nil {
		server.Storage.Del(key)
	}
	reply = ReplySwitch(err, BulkReply(value))
	return
}

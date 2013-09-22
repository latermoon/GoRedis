package goredis_server

import (
	. "../goredis"
	"./storage"
)

func (server *GoRedisServer) OnHGET(cmd *Command) (reply *Reply) {
	key := cmd.StringAtIndex(1)
	field := cmd.StringAtIndex(2)
	value, err := server.Storages.HashStorage.HGet(key, field)
	reply = ReplySwitch(err, BulkReply(value))
	return
}

func (server *GoRedisServer) OnHSET(cmd *Command) (reply *Reply) {
	key := cmd.StringAtIndex(1)
	field := cmd.StringAtIndex(2)
	value := cmd.StringAtIndex(3)
	result, err := server.Storages.HashStorage.HSet(key, field, value)
	server.Storages.KeyTypeStorage.SetType(key, storage.KeyTypeHash)
	reply = ReplySwitch(err, IntegerReply(result))
	return
}

func (server *GoRedisServer) OnHGETALL(cmd *Command) (reply *Reply) {
	key := cmd.StringAtIndex(1)
	keyvals, err := server.Storages.HashStorage.HGetAll(key)
	reply = ReplySwitch(err, MultiBulksReply(keyvals))
	return
}

func (server *GoRedisServer) OnHMGET(cmd *Command) (reply *Reply) {
	key := cmd.StringAtIndex(1)
	fields := cmd.StringArgs()[2:]
	values, err := server.Storages.HashStorage.HMGet(key, fields...)
	server.Storages.KeyTypeStorage.SetType(key, storage.KeyTypeHash)
	reply = ReplySwitch(err, MultiBulksReply(values))
	return
}

func (server *GoRedisServer) OnHMSET(cmd *Command) (reply *Reply) {
	key := cmd.StringAtIndex(1)
	keyvals := cmd.StringArgs()[2:]
	if len(keyvals)%2 != 0 {
		reply = ErrorReply("Bad field/value paires")
		return
	}
	err := server.Storages.HashStorage.HMSet(key, keyvals...)
	server.Storages.KeyTypeStorage.SetType(key, storage.KeyTypeHash)
	reply = ReplySwitch(err, StatusReply("OK"))
	return
}

func (server *GoRedisServer) OnHLEN(cmd *Command) (reply *Reply) {
	key := cmd.StringAtIndex(1)
	length, err := server.Storages.HashStorage.HLen(key)
	reply = ReplySwitch(err, IntegerReply(length))
	return
}

func (server *GoRedisServer) OnHDEL(cmd *Command) (reply *Reply) {
	key := cmd.StringAtIndex(1)
	fields := cmd.StringArgs()[2:]
	n, err := server.Storages.HashStorage.HDel(key, fields...)
	reply = ReplySwitch(err, IntegerReply(n))
	return
}

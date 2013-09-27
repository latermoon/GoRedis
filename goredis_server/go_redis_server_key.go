package goredis_server

import (
	. "../goredis"
	"./storage"
)

func (server *GoRedisServer) OnDEL(cmd *Command) (reply *Reply) {
	keys := cmd.StringArgs()[1:]
	count, _ := server.Storage.Del(keys...)
	reply = IntegerReply(count)
	return
}

func (server *GoRedisServer) OnTYPE(cmd *Command) (reply *Reply) {
	key := cmd.StringAtIndex(1)
	keytype := server.Storage.TypeOf(key)
	typestr := "none"
	switch keytype {
	case storage.KeyTypeString:
		typestr = "string"
	case storage.KeyTypeHash:
		typestr = "hash"
	case storage.KeyTypeList:
		typestr = "list"
	case storage.KeyTypeSet:
		typestr = "set"
	case storage.KeyTypeSortedSet:
		typestr = "zset"
	default:
		typestr = "none"
	}
	reply = StatusReply(typestr)
	return
}

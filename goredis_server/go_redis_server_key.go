package goredis_server

import (
	. "../goredis"
	. "./storage"
)

var typeTable = map[EntryType]string{EntryTypeString: "string", EntryTypeHash: "hash", EntryTypeList: "list", EntryTypeSet: "set", EntryTypeSortedSet: "zset"}

func (server *GoRedisServer) OnDEL(cmd *Command) (reply *Reply) {
	keys := cmd.StringArgs()[1:]
	n := 0
	for _, key := range keys {
		entry := server.datasource.Get(key)
		if entry != nil {
			server.datasource.Remove(key)
			n++
		}
	}
	reply = IntegerReply(n)
	return
}

func (server *GoRedisServer) OnTYPE(cmd *Command) (reply *Reply) {
	key := cmd.StringAtIndex(1)
	entry := server.datasource.Get(key)
	if entry != nil {
		if desc, exist := typeTable[entry.Type()]; exist {
			return StatusReply(desc)
		}
	}
	return StatusReply("none")
}

func (server *GoRedisServer) OnSAVEKEY(cmd *Command) (reply *Reply) {
	key := cmd.StringAtIndex(1)
	entry := server.datasource.Get(key)
	if entry == nil {
		return ErrorReply(key + " not exist")
	}
	err := server.datasource.Set(key, entry)
	reply = ReplySwitch(err, StatusReply("OK"))
	return
}

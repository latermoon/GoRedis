package goredis_server

import (
	. "../goredis"
	. "./storage"
)

func (server *GoRedisServer) OnDEL(cmd *Command) (reply *Reply) {
	keys := cmd.StringArgs()[1:]
	n := 0
	for _, key := range keys {
		entry, _ := server.datasource.Get(key)
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
	var keytype EntryType
	entry, _ := server.datasource.Get(key)
	if entry != nil {
		keytype = entry.Type()
	}
	var typestr string
	switch keytype {
	case EntryTypeString:
		typestr = "string"
	case EntryTypeHash:
		typestr = "hash"
	case EntryTypeList:
		typestr = "list"
	case EntryTypeSet:
		typestr = "set"
	case EntryTypeSortedSet:
		typestr = "zset"
	default:
		typestr = "none"
	}
	reply = StatusReply(typestr)
	return
}

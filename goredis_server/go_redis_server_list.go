package goredis_server

import (
	. "../goredis"
	. "./storage"
	"strconv"
)

func (server *GoRedisServer) OnLLEN(cmd *Command) (reply *Reply) {
	key := cmd.StringAtIndex(1)
	entry := server.datasource.Get(key)
	if entry == nil {
		reply = IntegerReply(0)
	} else if entry.Type() == EntryTypeList {
		length := entry.(*ListEntry).List().Len()
		reply = IntegerReply(length)
	} else {
		reply = WrongKindReply
	}
	return
}

func (server *GoRedisServer) OnLINDEX(cmd *Command) (reply *Reply) {
	key := cmd.StringAtIndex(1)
	index, _ := strconv.Atoi(cmd.StringAtIndex(2))
	entry := server.datasource.Get(key)
	if entry == nil {
		reply = BulkReply(nil)
	} else if entry.Type() == EntryTypeList {
		val := entry.(*ListEntry).List().Index(index)
		reply = BulkReply(val)
	} else {
		reply = WrongKindReply
	}
	return
}

/*
if entry == nil {

	} else if entry.Type() == EntryTypeList {
		entry.(*ListEntry).List()
	} else {
		reply = WrongKindReply
	}
*/
func (server *GoRedisServer) OnLRANGE(cmd *Command) (reply *Reply) {
	key := cmd.StringAtIndex(1)
	start, _ := strconv.Atoi(cmd.StringAtIndex(2))
	end, _ := strconv.Atoi(cmd.StringAtIndex(3))
	entry := server.datasource.Get(key)
	if entry == nil {
		reply = MultiBulksReply([]interface{}{})
	} else if entry.Type() == EntryTypeList {
		vals := entry.(*ListEntry).List().Range(start, end)
		reply = MultiBulksReply(vals)
	} else {
		reply = WrongKindReply
	}
	return
}

func (server *GoRedisServer) OnRPUSH(cmd *Command) (reply *Reply) {
	key := cmd.StringAtIndex(1)
	values := cmd.StringArgs()[2:]
	entry := server.datasource.Get(key)
	if entry == nil {
		entry = NewListEntry()
		server.datasource.Set(key, entry)
	} else if entry.Type() != EntryTypeList {
		reply = WrongKindReply
		return
	}
	n := entry.(*ListEntry).List().RPush(values...)
	reply = IntegerReply(n)
	return
}

func (server *GoRedisServer) OnLPOP(cmd *Command) (reply *Reply) {
	key := cmd.StringAtIndex(1)
	entry := server.datasource.Get(key)
	if entry == nil {
		reply = BulkReply(nil)
		return
	} else if entry.Type() != EntryTypeList {
		reply = WrongKindReply
		return
	}
	sl := entry.(*ListEntry).List()
	val := sl.LPop()
	if sl.Len() == 0 {
		server.datasource.Remove(key)
	}
	reply = BulkReply(val)
	return
}

func (server *GoRedisServer) OnLPUSH(cmd *Command) (reply *Reply) {
	key := cmd.StringAtIndex(1)
	values := cmd.StringArgs()[2:]
	entry := server.datasource.Get(key)
	if entry == nil {
		entry = NewListEntry()
		server.datasource.Set(key, entry)
	} else if entry.Type() != EntryTypeList {
		reply = WrongKindReply
		return
	}
	n := entry.(*ListEntry).List().LPush(values...)
	reply = IntegerReply(n)
	return
}

func (server *GoRedisServer) OnRPOP(cmd *Command) (reply *Reply) {
	key := cmd.StringAtIndex(1)
	entry := server.datasource.Get(key)
	if entry == nil {
		reply = BulkReply(nil)
		return
	} else if entry.Type() != EntryTypeList {
		reply = WrongKindReply
		return
	}
	sl := entry.(*ListEntry).List()
	val := sl.RPop()
	if sl.Len() == 0 {
		server.datasource.Remove(key)
	}
	reply = BulkReply(val)
	return
}

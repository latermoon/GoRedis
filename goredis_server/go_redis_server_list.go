package goredis_server

import (
	. "../goredis"
	. "./storage"
)

// 获取List，不存在则自动创建
func (server *GoRedisServer) listByKey(key []byte, create bool) (lst *ListEntry, err error) {
	entry := server.datasource.Get(key)
	if entry != nil && entry.Type() != EntryTypeList {
		err = WrongKindError
		return
	}
	if entry != nil {
		lst = entry.(*ListEntry)
	} else if create {
		lst = NewListEntry()
		server.datasource.Set(key, lst)
	}
	return
}

func (server *GoRedisServer) OnLLEN(cmd *Command) (reply *Reply) {
	key, _ := cmd.ArgAtIndex(1)
	entry, err := server.listByKey(key, false)
	if err != nil {
		return ErrorReply(err)
	} else if entry == nil {
		return IntegerReply(0)
	}
	n := entry.List().Len()
	reply = IntegerReply(n)
	return
}

func (server *GoRedisServer) OnLINDEX(cmd *Command) (reply *Reply) {
	key, _ := cmd.ArgAtIndex(1)
	entry, err := server.listByKey(key, false)
	if err != nil {
		return ErrorReply(err)
	} else if entry == nil {
		return BulkReply(nil)
	}

	index, e1 := cmd.IntAtIndex(2)
	if e1 != nil {
		return ErrorReply(e1)
	}
	val := entry.List().Index(index)
	reply = BulkReply(val)
	return
}

func (server *GoRedisServer) OnLRANGE(cmd *Command) (reply *Reply) {
	key, _ := cmd.ArgAtIndex(1)
	entry, err := server.listByKey(key, false)
	if err != nil {
		return ErrorReply(err)
	} else if entry == nil {
		return MultiBulksReply([]interface{}{})
	}

	start, e1 := cmd.IntAtIndex(2)
	end, e2 := cmd.IntAtIndex(3)
	if e1 != nil || e2 != nil {
		return ErrorReply("Bad start/end")
	}
	vals := entry.List().Range(start, end)
	reply = MultiBulksReply(vals)
	return
}

func (server *GoRedisServer) OnRPUSH(cmd *Command) (reply *Reply) {
	key, _ := cmd.ArgAtIndex(1)
	entry, err := server.listByKey(key, true)
	if err != nil {
		return ErrorReply(err)
	}

	values := cmd.Args[2:]
	objs := BytesToInterfaceSlice(values)
	n := entry.List().RPush(objs...)
	if n > 0 {
		server.datasource.NotifyEntryUpdate(key, entry)
	}
	reply = IntegerReply(n)
	return
}

func (server *GoRedisServer) OnLPOP(cmd *Command) (reply *Reply) {
	key, _ := cmd.ArgAtIndex(1)
	entry, err := server.listByKey(key, false)
	if err != nil {
		return ErrorReply(err)
	} else if entry == nil {
		return BulkReply(nil)
	}

	val := entry.List().LPop()
	if entry.List().Len() == 0 {
		server.datasource.Remove(key)
	} else {
		server.datasource.NotifyEntryUpdate(key, entry)
	}
	reply = BulkReply(val)
	return
}

func (server *GoRedisServer) OnLPUSH(cmd *Command) (reply *Reply) {
	key, _ := cmd.ArgAtIndex(1)
	entry, err := server.listByKey(key, true)
	if err != nil {
		return ErrorReply(err)
	}

	values := cmd.Args[2:]
	objs := BytesToInterfaceSlice(values)
	n := entry.List().LPush(objs...)
	if n > 0 {
		server.datasource.NotifyEntryUpdate(key, entry)
	}
	reply = IntegerReply(n)
	return
}

func (server *GoRedisServer) OnRPOP(cmd *Command) (reply *Reply) {
	key, _ := cmd.ArgAtIndex(1)
	entry, err := server.listByKey(key, false)
	if err != nil {
		return ErrorReply(err)
	} else if entry == nil {
		return BulkReply(nil)
	}

	val := entry.List().RPop()
	if entry.List().Len() == 0 {
		server.datasource.Remove(key)
	} else {
		server.datasource.NotifyEntryUpdate(key, entry)
	}
	reply = BulkReply(val)
	return
}

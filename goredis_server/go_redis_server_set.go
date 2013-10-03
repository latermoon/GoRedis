package goredis_server

import (
	. "../goredis"
	. "./storage"
)

// 获取SortedSet，不存在则自动创建
func (server *GoRedisServer) setByKey(key string) (se *SetEntry, err error) {
	entry := server.datasource.Get(key)
	if entry != nil && entry.Type() != EntryTypeSet {
		err = WrongKindError
		return
	}
	if entry == nil {
		entry = NewSetEntry()
		server.datasource.Set(key, entry)
	}
	se = entry.(*SetEntry)
	return
}

// SADD key member [member ...]
// Add one or more members to a set
func (server *GoRedisServer) OnSADD(cmd *Command) (reply *Reply) {
	key := cmd.StringAtIndex(1)
	entry, err := server.setByKey(key)
	if err != nil {
		return ErrorReply(err)
	}
	members := cmd.StringArgs()[2:]
	n := 0
	for _, member := range members {
		ok := entry.Put(member)
		if ok {
			n++
		}
	}
	if n > 0 {
		server.datasource.NotifyEntryUpdate(key, entry)
	}
	return IntegerReply(n)
}

func (server *GoRedisServer) OnSCARD(cmd *Command) (reply *Reply) {
	key := cmd.StringAtIndex(1)
	entry, err := server.setByKey(key)
	if err != nil {
		return ErrorReply(err)
	}
	n := entry.Count()
	return IntegerReply(n)
}

func (server *GoRedisServer) OnSISMEMBER(cmd *Command) (reply *Reply) {
	key := cmd.StringAtIndex(1)
	entry, err := server.setByKey(key)
	if err != nil {
		return ErrorReply(err)
	}
	member := cmd.StringAtIndex(2)
	if entry.Contains(member) {
		reply = IntegerReply(1)
	} else {
		reply = IntegerReply(0)
	}
	return
}

func (server *GoRedisServer) OnSMEMBERS(cmd *Command) (reply *Reply) {
	key := cmd.StringAtIndex(1)
	entry, err := server.setByKey(key)
	if err != nil {
		return ErrorReply(err)
	}
	keys := entry.Keys()
	reply = MultiBulksReply(keys)
	return
}

func (server *GoRedisServer) OnSREM(cmd *Command) (reply *Reply) {
	key := cmd.StringAtIndex(1)
	entry, err := server.setByKey(key)
	if err != nil {
		return ErrorReply(err)
	}
	members := cmd.StringArgs()[2:]
	n := 0
	for _, member := range members {
		ok := entry.Remove(member)
		if ok {
			n++
		}
	}
	if entry.Count() == 0 {
		server.datasource.Remove(key)
	} else {
		if n > 0 {
			server.datasource.NotifyEntryUpdate(key, entry)
		}
	}
	return IntegerReply(n)
}

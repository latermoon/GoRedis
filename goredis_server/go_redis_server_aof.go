package goredis_server

/*
自定义aof指令集，用于实现海量日志存储
aof_push key value [value ...]
aof_pop key 10
aof_index key index
aof_range key start end
aof_len key
*/

import (
	. "../goredis"
	"./libs/leveltool"
	. "./storage"
)

func (server *GoRedisServer) aoflistByKey(key string, create bool) (lst *leveltool.LevelList) {
	server.aoftableMutex.Lock()
	defer server.aoftableMutex.Unlock()

	var exist bool
	lst, exist = server.aoftable[key]
	if !exist {
		// inner key
		// __aof:user:100422:lochistory:_start = 0
		// __aof:user:100422:lochistory:_end = 1
		// __aof:user:100422:lochistory:idx:0 = hello
		// __aof:user:100422:lochistory:idx:1 = hello
		lst = leveltool.NewLevelList(server.datasource.(*LevelDBDataSource).DB(), "__goredis:aof:"+key)
		server.aoftable[key] = lst
	}
	return
}

func (server *GoRedisServer) OnAOF_PUSH(cmd *Command) (reply *Reply) {
	key := cmd.StringAtIndex(1)
	vals := cmd.Args[2:]
	lst := server.aoflistByKey(key, true)
	for _, val := range vals {
		lst.Push(val)
	}
	length := int(lst.Len())
	return IntegerReply(length)
}

func (server *GoRedisServer) OnAOF_POP(cmd *Command) (reply *Reply) {
	key := cmd.StringAtIndex(1)
	lst := server.aoflistByKey(key, false)
	if lst == nil {
		return BulkReply(nil)
	}
	elem, err := lst.Pop()
	if err != nil {
		return ErrorReply(err)
	}
	reply = BulkReply(elem.Value.([]byte))
	return
}

func (server *GoRedisServer) OnAOF_INDEX(cmd *Command) (reply *Reply) {
	key := cmd.StringAtIndex(1)
	lst := server.aoflistByKey(key, false)
	if lst == nil {
		return BulkReply(nil)
	}
	idx, err := cmd.IntAtIndex(2)
	if err != nil {
		return ErrorReply("bad index")
	}
	elem := lst.Element(int64(idx))
	if elem == nil {
		return BulkReply(nil)
	}
	reply = BulkReply(elem.Value.([]byte))
	return
}

// func (server *GoRedisServer) OnAOF_RANGE(cmd *Command) (reply *Reply) {
// 	key := cmd.StringAtIndex(1)
// 	return
// }

func (server *GoRedisServer) OnAOF_LEN(cmd *Command) (reply *Reply) {
	key := cmd.StringAtIndex(1)
	lst := server.aoflistByKey(key, false)
	if lst == nil {
		return IntegerReply(0)
	}
	n := int(lst.Len())
	reply = IntegerReply(n)
	return
}

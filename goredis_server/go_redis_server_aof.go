package goredis_server

/*
自定义aof指令集，用于实现海量日志存储
aof_push key value [value ...]    <IntegerReply: length>
aof_pop key    <BulkReply: nil>
aof_index key index    <BulkReply: nil>
aof_range key start end    <MultiBulksReply: nil>
aof_len key    <IntegerReply: 0>
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
		// 使用levellist实现
		lst = leveltool.NewLevelList(server.datasource.(*LevelDBDataSource).DB(), "__aof:"+key)
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
	elem, e2 := lst.Index(int64(idx))
	if e2 != nil {
		return ErrorReply(e2)
	} else if elem == nil {
		return BulkReply(nil)
	}
	reply = BulkReply(elem.Value.([]byte))
	return
}

func (server *GoRedisServer) OnAOF_RANGE(cmd *Command) (reply *Reply) {
	key := cmd.StringAtIndex(1)
	lst := server.aoflistByKey(key, false)
	if lst == nil {
		return MultiBulksReply([]interface{}{})
	}
	start, e1 := cmd.IntAtIndex(2)
	end, e2 := cmd.IntAtIndex(3)
	if e1 != nil || e2 != nil {
		return ErrorReply("bad start/end")
	} else if start < 0 {
		return ErrorReply("start > end")
	} else if end != -1 && start > end {
		return ErrorReply("start > end")
	}
	// 初始缓冲
	buflen := end - start + 1
	if end <= 0 || end > 100 {
		buflen = 100
	}
	bulks := make([]interface{}, 0, buflen)
	for i := start; end == -1 || i <= end; i++ {
		elem, e2 := lst.Index(int64(i))
		if e2 != nil {
			return ErrorReply(e2)
		} else if elem == nil {
			break
		}
		bulks = append(bulks, elem.Value.([]byte))
	}
	reply = MultiBulksReply(bulks)
	return
}

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

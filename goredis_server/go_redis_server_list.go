package goredis_server

import (
	. "../goredis"
)

func (server *GoRedisServer) OnLPUSH(cmd *Command) (reply *Reply) {
	key := cmd.StringAtIndex(1)
	vals := cmd.Args[2:]
	lst := server.keyManager.listByKey(key)
	for _, val := range vals {
		lst.LPush(val)
	}
	length := int(lst.Len())
	return IntegerReply(length)
}

func (server *GoRedisServer) OnRPUSH(cmd *Command) (reply *Reply) {
	key := cmd.StringAtIndex(1)
	vals := cmd.Args[2:]
	lst := server.keyManager.listByKey(key)
	for _, val := range vals {
		lst.RPush(val)
	}
	length := int(lst.Len())
	return IntegerReply(length)
}

func (server *GoRedisServer) OnRPOP(cmd *Command) (reply *Reply) {
	key := cmd.StringAtIndex(1)
	lst := server.keyManager.listByKey(key)
	if lst == nil {
		return BulkReply(nil)
	}
	elem, err := lst.RPop()
	if err != nil {
		return ErrorReply(err)
	}
	if elem != nil {
		reply = BulkReply(elem.Value.([]byte))
	} else {
		reply = BulkReply(nil)
	}
	return
}

func (server *GoRedisServer) OnLPOP(cmd *Command) (reply *Reply) {
	key := cmd.StringAtIndex(1)
	lst := server.keyManager.listByKey(key)
	if lst == nil {
		return BulkReply(nil)
	}
	elem, err := lst.LPop()
	if err != nil {
		return ErrorReply(err)
	}
	if elem != nil {
		reply = BulkReply(elem.Value.([]byte))
	} else {
		reply = BulkReply(nil)
	}
	return
}

func (server *GoRedisServer) OnLINDEX(cmd *Command) (reply *Reply) {
	key := cmd.StringAtIndex(1)
	lst := server.keyManager.listByKey(key)
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

func (server *GoRedisServer) OnLRANGE(cmd *Command) (reply *Reply) {
	key := cmd.StringAtIndex(1)
	lst := server.keyManager.listByKey(key)
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

func (server *GoRedisServer) OnLLEN(cmd *Command) (reply *Reply) {
	key := cmd.StringAtIndex(1)
	lst := server.keyManager.listByKey(key)
	if lst == nil {
		return IntegerReply(0)
	}
	n := int(lst.Len())
	reply = IntegerReply(n)
	return
}

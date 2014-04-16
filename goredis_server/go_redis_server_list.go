package goredis_server

import (
	. "GoRedis/goredis"
)

func (server *GoRedisServer) OnLPUSH(cmd *Command) (reply *Reply) {
	key := cmd.StringAtIndex(1)
	vals := cmd.Args()[2:]
	lst := server.levelRedis.GetList(key)
	err := lst.LPush(vals...)
	if err != nil {
		return ErrorReply(err)
	}
	length := int(lst.Len())
	return IntegerReply(length)
}

func (server *GoRedisServer) OnRPUSH(cmd *Command) (reply *Reply) {
	key := cmd.StringAtIndex(1)
	vals := cmd.Args()[2:]
	lst := server.levelRedis.GetList(key)
	err := lst.RPush(vals...)
	if err != nil {
		return ErrorReply(err)
	}
	length := int(lst.Len())
	return IntegerReply(length)
}

func (server *GoRedisServer) OnRPOP(cmd *Command) (reply *Reply) {
	key := cmd.StringAtIndex(1)
	lst := server.levelRedis.GetList(key)
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
	lst := server.levelRedis.GetList(key)
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
	lst := server.levelRedis.GetList(key)
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

func (server *GoRedisServer) OnLTRIM(cmd *Command) (reply *Reply) {
	key := cmd.StringAtIndex(1)
	lst := server.levelRedis.GetList(key)
	start, e1 := cmd.IntAtIndex(2)
	end, e2 := cmd.IntAtIndex(3)
	if e1 != nil || e2 != nil {
		return ErrorReply("bad start/stop")
	} else if start != 0 {
		return ErrorReply("start must equal to 0 (in GoRedis)")
	} else if end < start {
		return ErrorReply("end must greater than start")
	}
	lst.TrimLeft(uint(end - start + 1))
	reply = StatusReply("OK")
	return
}

func (server *GoRedisServer) OnLRANGE(cmd *Command) (reply *Reply) {
	key := cmd.StringAtIndex(1)
	start, e1 := cmd.Int64AtIndex(2)
	end, e2 := cmd.Int64AtIndex(3)
	if e1 != nil || e2 != nil {
		return ErrorReply("bad start/end")
	} else if start < 0 {
		return ErrorReply("start > end")
	} else if end != -1 && start > end {
		return ErrorReply("start > end")
	}

	lst := server.levelRedis.GetList(key)
	elems, err := lst.Range(start, end)
	if err != nil {
		return ErrorReply(err)
	}

	bulks := make([]interface{}, len(elems))
	for i := 0; i < len(elems); i++ {
		bulks[i] = elems[i].Value
	}
	reply = MultiBulksReply(bulks)
	return
}

func (server *GoRedisServer) OnLLEN(cmd *Command) (reply *Reply) {
	key := cmd.StringAtIndex(1)
	lst := server.levelRedis.GetList(key)
	if lst == nil {
		return IntegerReply(0)
	}
	n := int(lst.Len())
	reply = IntegerReply(n)
	return
}

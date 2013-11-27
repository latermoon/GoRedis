package goredis_server

import (
	. "../goredis"
)

// 在数据量大的情况下，keys基本不可用，使用key_search来分段扫描全部key
func (server *GoRedisServer) OnKEYS(cmd *Command) (reply *Reply) {
	return ErrorReply("keys is not supported by GoRedis, use 'key_search [prefix] [count] [withtype]' instead")
}

// 找出下一个key
// @return ["user:100422:name", "string", "user:100428:name", "string", "user:100422:setting", "hash", ...]
func (server *GoRedisServer) OnKEY_SEARCH(cmd *Command) (reply *Reply) {
	seekkey, err := cmd.ArgAtIndex(1)
	if err != nil {
		return ErrorReply(err)
	}
	count := 1
	if len(cmd.Args) > 2 {
		count, err = cmd.IntAtIndex(2)
		if err != nil {
			return ErrorReply(err)
		}
		if count < 1 || count > 10000 {
			return ErrorReply("count range: 1 < count < 10000")
		}
	}
	withtype := false
	if len(cmd.Args) > 3 {
		withtype = cmd.StringAtIndex(3) == "withtype"
	}
	// search
	bulks := server.keyManager.levelKey().Search(seekkey, "next", count, withtype, false)
	return MultiBulksReply(bulks)
}

func (server *GoRedisServer) OnKEY_SEARCH_DEL(cmd *Command) (reply *Reply) {
	seekkey, err := cmd.ArgAtIndex(1)
	if err != nil {
		return ErrorReply(err)
	}
	n := 0
	for {
		keys := server.keyManager.levelKey().Search(seekkey, "next", 1000, false, false)
		if len(keys) == 0 {
			break
		}
		for _, key := range keys {
			n += server.keyManager.Delete(key.([]byte))
		}
	}
	reply = IntegerReply(n)
	return
}

// 扫描内部key
func (server *GoRedisServer) OnGOKEY_SEARCH(cmd *Command) (reply *Reply) {
	seekkey, err := cmd.ArgAtIndex(1)
	if err != nil {
		return ErrorReply(err)
	}
	count := 1
	if len(cmd.Args) > 2 {
		count, err = cmd.IntAtIndex(2)
		if err != nil {
			return ErrorReply(err)
		}
		if count < 1 || count > 10000 {
			return ErrorReply("count range: 1 < count < 10000")
		}
	}
	withtype := false
	if len(cmd.Args) > 3 {
		withtype = cmd.StringAtIndex(3) == "withtype"
	}
	// search
	bulks := server.keyManager.levelKey().Search(seekkey, "next", count, withtype, true)
	return MultiBulksReply(bulks)
}

// 获取原始内容
func (server *GoRedisServer) OnGOGET(cmd *Command) (reply *Reply) {
	key, _ := cmd.ArgAtIndex(1)
	value := server.keyManager.levelKey().GetInnerValue(key)
	if value == nil {
		reply = BulkReply(nil)
	} else {
		reply = BulkReply(value)
	}
	return
}

func (server *GoRedisServer) OnDEL(cmd *Command) (reply *Reply) {
	keys := cmd.Args[1:]
	n := server.keyManager.Delete(keys...)
	reply = IntegerReply(n)
	return
}

func (server *GoRedisServer) OnTYPE(cmd *Command) (reply *Reply) {
	key, _ := cmd.ArgAtIndex(1)
	t := server.keyManager.levelKey().TypeOf(key)
	if len(t) > 0 {
		reply = StatusReply(t)
	} else {
		reply = StatusReply("none")
	}
	return
}

package goredis_server

/*
keysearch [prefix] [count]
keynext [min] [count]
*/

import (
	. "GoRedis/goredis"
	"GoRedis/libs/levelredis"
)

func (server *GoRedisServer) OnPING(cmd *Command) (reply *Reply) {
	reply = StatusReply("PONG")
	return
}

// 命名上使用keys来提供keysearch的扫描功能比较合理，
// 但是为了防止开发人员错误把适用与GoRedis的keys处理代码误用到官方redis上造成卡死
// 这里还是把keys禁用
func (server *GoRedisServer) OnKEYS(cmd *Command) (reply *Reply) {
	return ErrorReply("keys is not supported by GoRedis, use 'keysearch [prefix] [count] [withtype]' instead")
}

// 找出下一个key
func (server *GoRedisServer) OnKEYSEARCH(cmd *Command) (reply *Reply) {
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
	bulks := make([]interface{}, 0, 10)
	server.levelRedis.Keys(seekkey, func(i int, key, keytype []byte, quit *bool) {
		bulks = append(bulks, key)
		if withtype {
			bulks = append(bulks, keytype)
		}
		if i >= count-1 {
			*quit = true
		}
	})
	return MultiBulksReply(bulks)
}

// 扫描内部key
func (server *GoRedisServer) OnRAW_KEYSEARCH(cmd *Command) (reply *Reply) {
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
	// search
	bulks := make([]interface{}, 0, 10)
	min := seekkey
	max := append(seekkey, 254)
	server.levelRedis.Enumerate(min, max, levelredis.IterForward, func(i int, key, value []byte, quit *bool) {
		bulks = append(bulks, key)
		if i >= count-1 {
			*quit = true
		}
	})
	return MultiBulksReply(bulks)
}

// 获取原始内容
func (server *GoRedisServer) OnRAW_GET(cmd *Command) (reply *Reply) {
	key, _ := cmd.ArgAtIndex(1)
	value := server.levelRedis.RawGet(key)
	if value == nil {
		reply = BulkReply(nil)
	} else {
		reply = BulkReply(value)
	}
	return
}

/**
 * 过期时间，暂不支持
 * 1 if the timeout was set.
 * 0 if key does not exist or the timeout could not be set.
 */
func (server *GoRedisServer) OnEXPIRE(cmd *Command) (reply *Reply) {
	reply = IntegerReply(0)
	return
}

func (server *GoRedisServer) OnDEL(cmd *Command) (reply *Reply) {
	keys := cmd.Args[1:]
	n := server.levelRedis.Delete(keys...)
	reply = IntegerReply(n)
	return
}

func (server *GoRedisServer) OnTYPE(cmd *Command) (reply *Reply) {
	key, _ := cmd.ArgAtIndex(1)
	t := server.levelRedis.TypeOf(key)
	if len(t) > 0 {
		reply = StatusReply(t)
	} else {
		reply = StatusReply("none")
	}
	return
}

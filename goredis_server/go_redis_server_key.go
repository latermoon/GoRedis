package goredis_server

import (
	. "GoRedis/goredis"
	"GoRedis/libs/levelredis"
	"strings"
)

func (server *GoRedisServer) OnPING(cmd Command) (reply Reply) {
	reply = StatusReply("PONG")
	return
}

func (server *GoRedisServer) OnKEYS(cmd Command) (reply Reply) {
	return ErrorReply("use 'KEYSERACH [PREFIX] [LIMIT]' instead")
}

// keys重命名为keysearch
func (server *GoRedisServer) OnKEYSEARCH(cmd Command) (reply Reply) {
	seekkey := []byte("")
	if cmd.Len() > 1 {
		seekkey = cmd.Args()[1]
	}

	count := 10
	if cmd.Len() > 2 {
		var err error
		if count, err = cmd.IntAtIndex(2); err != nil {
			return ErrorReply(err)
		}
		if count < 1 || count > 10000 {
			return ErrorReply("count range: 1 < count < 10000")
		}
	}

	withtype := false
	if cmd.Len() > 3 {
		withtype = strings.ToUpper(cmd.StringAtIndex(3)) == "WITHTYPE"
	}

	// search
	bulks := make([]interface{}, 0, count)
	server.levelRedis.Keys(seekkey, func(i int, key, keytype []byte, quit *bool) {
		if i >= count {
			*quit = true
			return
		}
		bulks = append(bulks, key)
		if withtype {
			bulks = append(bulks, keytype)
		}
	})
	return MultiBulksReply(bulks)
}

// keyprev [seek] [count] [withtype] [withvalue]
func (server *GoRedisServer) OnKEYPREV(cmd Command) (reply Reply) {
	return server.keyEnumerate(cmd, levelredis.IterBackward)
}

// keynext [seek] [count] [withtype] [withvalue]
// 1) [key]
// 2) [type]
// 3) [value]
// 4) [key2]
// 5) [type2]
// 6) [value2]
func (server *GoRedisServer) OnKEYNEXT(cmd Command) (reply Reply) {
	return server.keyEnumerate(cmd, levelredis.IterForward)
}

func (server *GoRedisServer) keyEnumerate(cmd Command, direction levelredis.IterDirection) (reply Reply) {
	seek := cmd.Args()[1]
	count := 1
	withtype := false
	withvalue := false
	argcount := len(cmd.Args())
	if argcount > 2 {
		var err error
		count, err = cmd.IntAtIndex(2)
		if err != nil {
			return ErrorReply(err)
		}
		if count < 1 || count > 10000 {
			return ErrorReply("count range: 1 < count < 10000")
		}
	}
	if argcount > 3 {
		withtype = strings.ToUpper(cmd.StringAtIndex(3)) == "WITHTYPE"
	}
	// 必须withtype才能withvalue
	if withtype && argcount > 4 {
		withvalue = strings.ToUpper(cmd.StringAtIndex(4)) == "WITHVALUE"
	}
	// bulks初始大小
	bufferSize := count
	if withtype {
		bufferSize = count * 2
		if withvalue {
			bufferSize = count * 3
		}
	}
	bulks := make([]interface{}, 0, bufferSize)
	server.levelRedis.KeyEnumerate(seek, direction, func(i int, key, keytype, value []byte, quit *bool) {
		// stdlog.Println(i, string(key), string(keytype), string(value))
		bulks = append(bulks, key)
		if withtype {
			bulks = append(bulks, keytype)
			if withvalue {
				bulks = append(bulks, value)
			}
		}
		if i >= count-1 {
			*quit = true
		}
	})
	return MultiBulksReply(bulks)
}

// 扫描内部key
func (server *GoRedisServer) OnRAW_KEYSEARCH(cmd Command) (reply Reply) {
	seekkey := []byte("")
	if cmd.Len() > 1 {
		seekkey = cmd.Args()[1]
	}

	count := 10
	if cmd.Len() > 2 {
		var err error
		if count, err = cmd.IntAtIndex(2); err != nil {
			return ErrorReply(err)
		}
		if count < 1 || count > 10000 {
			return ErrorReply("count range: 1 < count < 10000")
		}
	}

	// search
	bulks := make([]interface{}, 0, count)
	min := seekkey
	max := append(seekkey, levelredis.MAXBYTE)
	server.levelRedis.RangeEnumerate(min, max, levelredis.IterForward, func(i int, key, value []byte, quit *bool) {
		bulks = append(bulks, key)
		if i >= count-1 {
			*quit = true
		}
	})
	return MultiBulksReply(bulks)
}

// 操作原始内容
func (server *GoRedisServer) OnRAW_GET(cmd Command) (reply Reply) {
	key, _ := cmd.ArgAtIndex(1)
	value, err := server.levelRedis.RawGet(key)
	if err != nil {
		return ErrorReply(err)
	}
	if value == nil {
		return BulkReply(nil)
	} else {
		return BulkReply(value)
	}
}

// 操作原始内容 RAW_SET +[hash]name latermoon
func (server *GoRedisServer) OnRAW_SET(cmd Command) (reply Reply) {
	key, value := cmd.Args()[1], cmd.Args()[2]
	err := server.levelRedis.RawSet(key, value)
	if err != nil {
		return ErrorReply(err)
	} else {
		return StatusReply("OK")
	}
}

// 官方redis的dbsize输出key数量，这里输出数据库大小
func (server *GoRedisServer) OnDBSIZE(cmd Command) (reply Reply) {
	return StatusReply(bytesInHuman(server.info.db_size()))
}

/**
 * 过期时间，暂不支持
 * 1 if the timeout was set.
 * 0 if key does not exist or the timeout could not be set.
 */
func (server *GoRedisServer) OnEXPIRE(cmd Command) (reply Reply) {
	reply = IntegerReply(0)
	return
}

func (server *GoRedisServer) OnDEL(cmd Command) (reply Reply) {
	keys := cmd.Args()[1:]
	n := server.levelRedis.Delete(keys...)
	reply = IntegerReply(n)
	return
}

func (server *GoRedisServer) OnTYPE(cmd Command) (reply Reply) {
	key, _ := cmd.ArgAtIndex(1)
	t := server.levelRedis.TypeOf(key)
	if len(t) > 0 {
		reply = StatusReply(t)
	} else {
		reply = StatusReply("none")
	}
	return
}

package goredis_server

// TODO 严谨的情况下应该校验参数数量，这里大部分都不校验是为了简化代码，panic后会断开client connection

import (
	. "GoRedis/goredis"
	"strconv"
)

var maxCmdLock = 100

func (server *GoRedisServer) OnGET(cmd *Command) (reply *Reply) {
	key := cmd.Args[1]
	value := server.levelRedis.Strings().Get(key)
	return BulkReply(value)
}

func (server *GoRedisServer) OnSET(cmd *Command) (reply *Reply) {
	key, _ := cmd.ArgAtIndex(1)
	val, _ := cmd.ArgAtIndex(2)
	server.levelRedis.Strings().Set(key, val)
	return StatusReply("OK")
}

func (server *GoRedisServer) OnMGET(cmd *Command) (reply *Reply) {
	keys := cmd.Args[1:]
	vals := make([]interface{}, len(keys))
	for i, key := range keys {
		vals[i] = server.levelRedis.Strings().Get(key)
	}
	reply = MultiBulksReply(vals)
	return
}

func (server *GoRedisServer) OnMSET(cmd *Command) (reply *Reply) {
	keyvals := cmd.Args[1:]
	if len(keyvals)%2 != 0 {
		return ErrorReply(WrongArgumentCount)
	}
	for i, count := 0, cmd.Len(); i < count; i += 2 {
		key := keyvals[i]
		val := keyvals[i+1]
		server.levelRedis.Strings().Set(key, val)
	}
	return StatusReply("OK")
}

/**
 * 计数器基于字符串，对字符串进行修改
 * TODO 性能需要改进
 * @param chg 增减量，正负数均可
 */
func (server *GoRedisServer) incrStringEntry(key []byte, chg int) (newvalue int, err error) {
	// 对操作的key进行hash后，有序并发处理
	hash := inthash(key, maxCmdLock)
	mu := mutexof("cmd_lock_" + strconv.Itoa(hash))
	mu.Lock()
	defer mu.Unlock()

	value := server.levelRedis.Strings().Get(key)
	var oldvalue int
	if value == nil {
		oldvalue = 0
	} else {
		oldvalue, err = strconv.Atoi(string(value))
		if err != nil {
			return
		}
	}
	// update
	newvalue = oldvalue + chg
	err = server.levelRedis.Strings().Set(key, []byte(strconv.Itoa(newvalue)))
	return
}

func (server *GoRedisServer) OnINCR(cmd *Command) (reply *Reply) {
	key, _ := cmd.ArgAtIndex(1)
	newvalue, err := server.incrStringEntry(key, 1)
	if err != nil {
		reply = ErrorReply(err)
	} else {
		reply = IntegerReply(newvalue)
	}
	return
}

func (server *GoRedisServer) OnINCRBY(cmd *Command) (reply *Reply) {
	key, _ := cmd.ArgAtIndex(1)
	chg, e1 := strconv.Atoi(cmd.StringAtIndex(2))
	if e1 != nil {
		reply = ErrorReply(e1)
		return
	}
	newvalue, err := server.incrStringEntry(key, chg)
	if err != nil {
		reply = ErrorReply(err)
	} else {
		reply = IntegerReply(newvalue)
	}
	return
}

func (server *GoRedisServer) OnDECR(cmd *Command) (reply *Reply) {
	key, _ := cmd.ArgAtIndex(1)
	newvalue, err := server.incrStringEntry(key, -1)
	if err != nil {
		reply = ErrorReply(err)
	} else {
		reply = IntegerReply(newvalue)
	}
	return
}

func (server *GoRedisServer) OnDECRBY(cmd *Command) (reply *Reply) {
	key, _ := cmd.ArgAtIndex(1)
	chg, e1 := strconv.Atoi(cmd.StringAtIndex(2))
	if e1 != nil {
		reply = ErrorReply(e1)
		return
	}
	newvalue, err := server.incrStringEntry(key, chg*-1)
	if err != nil {
		reply = ErrorReply(err)
	} else {
		reply = IntegerReply(newvalue)
	}
	return
}

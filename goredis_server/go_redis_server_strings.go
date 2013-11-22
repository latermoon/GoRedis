package goredis_server

// TODO 严谨的情况下应该校验参数数量，这里大部分都不校验是为了简化代码，panic后会断开client connection

import (
	. "../goredis"
	"strconv"
)

func (server *GoRedisServer) OnGET(cmd *Command) (reply *Reply) {
	key, _ := cmd.ArgAtIndex(1)
	value := server.levelString.Get(key)
	if value == nil {
		reply = BulkReply(nil)
	} else {
		reply = BulkReply(value)
	}
	return
}

func (server *GoRedisServer) OnSET(cmd *Command) (reply *Reply) {
	key, _ := cmd.ArgAtIndex(1)
	val, _ := cmd.ArgAtIndex(2)
	err := server.levelString.Set(key, val)
	reply = ReplySwitch(err, StatusReply("OK"))
	return
}

func (server *GoRedisServer) OnMGET(cmd *Command) (reply *Reply) {
	keys := cmd.Args[1:]
	vals := make([]interface{}, len(keys))
	for i, key := range keys {
		value := server.levelString.Get(key)
		if value == nil {
			vals[i] = nil
		} else {
			vals[i] = value
		}
	}
	reply = MultiBulksReply(vals)
	return
}

func (server *GoRedisServer) OnMSET(cmd *Command) (reply *Reply) {
	keyvals := cmd.Args[1:]
	count := len(keyvals)
	if count%2 != 0 {
		return ErrorReply("Bad Argument Count")
	}
	for i := 0; i < count; i += 2 {
		key := keyvals[i]
		val := keyvals[i+1]
		server.levelString.Set(key, val)
	}
	reply = StatusReply("OK")
	return
}

/**
 * 计数器基于字符串，对字符串进行修改
 * TODO 性能需要改进
 * @param chg 增减量，正负数均可
 */
func (server *GoRedisServer) incrStringEntry(key []byte, chg int) (newvalue int, err error) {
	value := server.levelString.Get(key)
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
	err = server.levelString.Set(key, []byte(strconv.Itoa(newvalue)))
	return
}

func (server *GoRedisServer) OnINCR(cmd *Command) (reply *Reply) {
	server.stringMutex.Lock()
	defer server.stringMutex.Unlock()

	key, _ := cmd.ArgAtIndex(1)
	newvalue, err := server.incrStringEntry(key, 1)
	reply = ReplySwitch(err, IntegerReply(newvalue))
	return
}

func (server *GoRedisServer) OnINCRBY(cmd *Command) (reply *Reply) {
	server.stringMutex.Lock()
	defer server.stringMutex.Unlock()

	key, _ := cmd.ArgAtIndex(1)
	chg, e1 := strconv.Atoi(cmd.StringAtIndex(2))
	if e1 != nil {
		reply = ErrorReply(e1)
		return
	}
	newvalue, err := server.incrStringEntry(key, chg)
	reply = ReplySwitch(err, IntegerReply(newvalue))
	return
}

func (server *GoRedisServer) OnDECR(cmd *Command) (reply *Reply) {
	server.stringMutex.Lock()
	defer server.stringMutex.Unlock()

	key, _ := cmd.ArgAtIndex(1)
	newvalue, err := server.incrStringEntry(key, -1)
	reply = ReplySwitch(err, IntegerReply(newvalue))
	return
}

func (server *GoRedisServer) OnDECRBY(cmd *Command) (reply *Reply) {
	server.stringMutex.Lock()
	defer server.stringMutex.Unlock()

	key, _ := cmd.ArgAtIndex(1)
	chg, e1 := strconv.Atoi(cmd.StringAtIndex(2))
	if e1 != nil {
		reply = ErrorReply(e1)
		return
	}
	newvalue, err := server.incrStringEntry(key, chg*-1)
	reply = ReplySwitch(err, IntegerReply(newvalue))
	return
}

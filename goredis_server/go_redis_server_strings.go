package goredis_server

// TODO 严谨的情况下应该校验参数数量，这里大部分都不校验是为了简化代码，panic后会断开client connection

import (
	. "../goredis"
	. "./storage"
	"strconv"
)

// 获取String，不存在则自动创建
func (server *GoRedisServer) stringByKey(key string, create bool) (e *StringEntry, err error) {
	entry := server.datasource.Get(key)
	if entry != nil && entry.Type() != EntryTypeString {
		err = WrongKindError
		return
	}
	if entry != nil {
		e = entry.(*StringEntry)
	} else if create {
		e = NewStringEntry(nil)
		server.datasource.Set(key, e)
	}
	return
}

func (server *GoRedisServer) OnGET(cmd *Command) (reply *Reply) {
	key := cmd.StringAtIndex(1)
	entry, err := server.stringByKey(key, false)
	if err != nil {
		return ErrorReply(err)
	} else if entry == nil {
		return BulkReply(nil)
	}
	reply = BulkReply(entry.Value())
	return
}

func (server *GoRedisServer) OnSET(cmd *Command) (reply *Reply) {
	key := cmd.StringAtIndex(1)
	val := cmd.StringAtIndex(2)
	entry := NewStringEntry(val)
	err := server.datasource.Set(key, entry)
	reply = ReplySwitch(err, StatusReply("OK"))
	return
}

func (server *GoRedisServer) OnMGET(cmd *Command) (reply *Reply) {
	keys := cmd.StringArgs()[1:]
	vals := make([]interface{}, len(keys))
	for i, key := range keys {
		entry, err := server.stringByKey(key, false)
		if err != nil || entry == nil {
			vals[i] = nil
			continue
		} else {
			vals[i] = entry.Value()
		}
	}
	reply = MultiBulksReply(vals)
	return
}

func (server *GoRedisServer) OnMSET(cmd *Command) (reply *Reply) {
	keyvals := cmd.StringArgs()[1:]
	count := len(keyvals)
	if count%2 != 0 {
		return ErrorReply("Bad Argument Count")
	}
	for i := 0; i < count; i += 2 {
		key := keyvals[i]
		val := keyvals[i+1]
		// 不做类型检查，其他数据会被string覆盖
		entry := NewStringEntry(val)
		server.datasource.Set(key, entry)
	}
	reply = StatusReply("OK")
	return
}

/**
 * 计数器基于字符串，对字符串进行修改
 * TODO 性能需要改进
 * @param chg 增减量，正负数均可
 */
func (server *GoRedisServer) incrStringEntry(key string, chg int) (newvalue int, err error) {
	entry := server.datasource.Get(key)
	if entry == nil {
		entry = NewStringEntry("0")
	} else if entry.Type() != EntryTypeString {
		err = WrongKindError
		return
	}
	// incr
	var oldvalue int
	oldvalue, err = strconv.Atoi(entry.(*StringEntry).String())
	if err != nil {
		return
	}
	newvalue = oldvalue + chg
	entry.(*StringEntry).SetValue(strconv.Itoa(newvalue))
	server.datasource.Set(key, entry)
	return
}

func (server *GoRedisServer) OnINCR(cmd *Command) (reply *Reply) {
	server.stringMutex.Lock()
	defer server.stringMutex.Unlock()

	key := cmd.StringAtIndex(1)
	newvalue, err := server.incrStringEntry(key, 1)
	reply = ReplySwitch(err, IntegerReply(newvalue))
	return
}

func (server *GoRedisServer) OnINCRBY(cmd *Command) (reply *Reply) {
	server.stringMutex.Lock()
	defer server.stringMutex.Unlock()

	key := cmd.StringAtIndex(1)
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

	key := cmd.StringAtIndex(1)
	newvalue, err := server.incrStringEntry(key, -1)
	reply = ReplySwitch(err, IntegerReply(newvalue))
	return
}

func (server *GoRedisServer) OnDECRBY(cmd *Command) (reply *Reply) {
	server.stringMutex.Lock()
	defer server.stringMutex.Unlock()

	key := cmd.StringAtIndex(1)
	chg, e1 := strconv.Atoi(cmd.StringAtIndex(2))
	if e1 != nil {
		reply = ErrorReply(e1)
		return
	}
	newvalue, err := server.incrStringEntry(key, chg*-1)
	reply = ReplySwitch(err, IntegerReply(newvalue))
	return
}

package goredis_server

import (
	. "../goredis"
	. "./storage"
	"strconv"
)

func (server *GoRedisServer) OnGET(cmd *Command) (reply *Reply) {
	// [TODO] 严谨的情况下应该校验参数数量，这里大部分都不校验是为了简化代码，panic后会断开client connection
	key := cmd.StringAtIndex(1)
	entry := server.datasource.Get(key)
	if entry == nil {
		reply = BulkReply(nil)
	} else if entry.Type() == EntryTypeString {
		reply = BulkReply(entry.(*StringEntry).Value())
	} else {
		reply = WrongKindReply
	}

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
		entry := server.datasource.Get(key)
		if entry != nil && entry.Type() == EntryTypeString {
			vals[i] = entry.(*StringEntry).Value()
		} else {
			vals[i] = nil
		}
	}
	reply = MultiBulksReply(vals)
	return
}

func (server *GoRedisServer) OnMSET(cmd *Command) (reply *Reply) {
	// TODO 是否需要加lock
	keyvals := cmd.StringArgs()[1:]
	count := len(keyvals)
	if count%2 != 0 {
		return ErrorReply("Bad Argument Count")
	}
	for i := 0; i < count; i += 2 {
		key := keyvals[i]
		val := keyvals[i+1]
		entry := NewStringEntry(val)
		server.datasource.Set(key, entry)
	}
	reply = StatusReply("OK")
	return
}

/**
 * 计数器基于字符串，对字符串进行修改
 * TODO 性能需要改进
 * @param count 正负数均可
 */
func (server *GoRedisServer) incrValue(entry *StringEntry, count int) (newvalue int, err error) {
	var oldvalue int
	oldvalue, err = strconv.Atoi(entry.String())
	if err != nil {
		return
	}
	newvalue = oldvalue + count
	entry.SetValue(strconv.Itoa(newvalue))
	return
}

func (server *GoRedisServer) OnINCR(cmd *Command) (reply *Reply) {
	server.stringMutex.Lock()
	defer server.stringMutex.Unlock()

	key := cmd.StringAtIndex(1)
	entry := server.datasource.Get(key)
	if entry == nil {
		entry = NewStringEntry("0")
	} else if entry.Type() != EntryTypeString {
		reply = WrongKindReply
		return
	}
	// incr
	newvalue, err := server.incrValue(entry.(*StringEntry), 1)
	if err != nil {
		reply = ErrorReply(err)
		return
	}
	server.datasource.Set(key, entry)

	reply = IntegerReply(newvalue)
	return
}

func (server *GoRedisServer) OnDECR(cmd *Command) (reply *Reply) {
	server.stringMutex.Lock()
	defer server.stringMutex.Unlock()

	key := cmd.StringAtIndex(1)
	entry := server.datasource.Get(key)
	if entry == nil {
		entry = NewStringEntry("0")
	} else if entry.Type() != EntryTypeString {
		reply = WrongKindReply
		return
	}
	// decr
	newvalue, err := server.incrValue(entry.(*StringEntry), -1)
	if err != nil {
		reply = ErrorReply(err)
		return
	}
	server.datasource.Set(key, entry)

	reply = IntegerReply(newvalue)
	return
}

package goredis_server

import (
	. "../goredis"
	. "./storage"
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

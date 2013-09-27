package goredis_server

import (
	. "../goredis"
)

func (server *GoRedisServer) OnGET(cmd *Command) (reply *Reply) {
	// [TODO] 严谨的情况下应该校验参数数量，这里大部分都不校验是为了简化代码，panic后会断开client connection
	key := cmd.StringAtIndex(1)
	val, err := server.Storage.Get(key)
	reply = ReplySwitch(err, BulkReply(val))
	return
}

func (server *GoRedisServer) OnSET(cmd *Command) (reply *Reply) {
	key := cmd.StringAtIndex(1)
	val := cmd.StringAtIndex(2)
	err := server.Storage.Set(key, val)
	reply = ReplySwitch(err, StatusReply("OK"))
	return
}

func (server *GoRedisServer) OnMGET(cmd *Command) (reply *Reply) {
	keys := cmd.StringArgs()[1:]
	vals, err := server.Storage.MGet(keys...)
	reply = ReplySwitch(err, MultiBulksReply(vals))
	return
}

func (server *GoRedisServer) OnMSET(cmd *Command) (reply *Reply) {
	keyvals := cmd.StringArgs()[1:]
	if len(keyvals)%2 != 0 {
		return ErrorReply("Bad Argument Count")
	}
	err := server.Storage.MSet(keyvals...)
	reply = ReplySwitch(err, StatusReply("OK"))
	return
}

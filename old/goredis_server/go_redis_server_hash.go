package goredis_server

import (
	. "GoRedis/goredis"
)

func (server *GoRedisServer) OnHGET(cmd *Command) (reply *Reply) {
	key := cmd.StringAtIndex(1)
	field, _ := cmd.ArgAtIndex(2)
	hash := server.levelRedis.GetHash(key)
	val := hash.Get(field)
	if val == nil {
		reply = BulkReply(nil)
	} else {
		reply = BulkReply(val)
	}
	return
}

func (server *GoRedisServer) OnHSET(cmd *Command) (reply *Reply) {
	key := cmd.StringAtIndex(1)
	hash := server.levelRedis.GetHash(key)
	field, _ := cmd.ArgAtIndex(2)
	value, _ := cmd.ArgAtIndex(3)
	hash.Set(field, value)
	return IntegerReply(1)
}

func (server *GoRedisServer) OnHGETALL(cmd *Command) (reply *Reply) {
	key := cmd.StringAtIndex(1)
	hash := server.levelRedis.GetHash(key)
	elems := hash.GetAll(1000)
	keyvals := make([]interface{}, 0, len(elems)*2)
	for _, elem := range elems {
		keyvals = append(keyvals, elem.Key)
		keyvals = append(keyvals, elem.Value)
	}
	reply = MultiBulksReply(keyvals)
	return
}

func (server *GoRedisServer) OnHMGET(cmd *Command) (reply *Reply) {
	key := cmd.StringAtIndex(1)
	hash := server.levelRedis.GetHash(key)
	fields := cmd.Args()[2:]
	keyvals := make([]interface{}, 0, len(fields))
	for _, field := range fields {
		val := hash.Get(field)
		keyvals = append(keyvals, val)
	}
	reply = MultiBulksReply(keyvals)
	return
}

func (server *GoRedisServer) OnHMSET(cmd *Command) (reply *Reply) {
	key := cmd.StringAtIndex(1)
	keyvals := cmd.Args()[2:]
	if len(keyvals)%2 != 0 {
		reply = ErrorReply("Bad field/value paires")
		return
	}
	hash := server.levelRedis.GetHash(key)
	hash.Set(keyvals...)
	reply = StatusReply("OK")
	return
}

func (server *GoRedisServer) OnHEXISTS(cmd *Command) (reply *Reply) {
	key := cmd.StringAtIndex(1)
	field, _ := cmd.ArgAtIndex(2)
	hash := server.levelRedis.GetHash(key)
	val := hash.Get(field)
	if val == nil {
		reply = IntegerReply(0)
	} else {
		reply = IntegerReply(1)
	}
	return
}

func (server *GoRedisServer) OnHLEN(cmd *Command) (reply *Reply) {
	key := cmd.StringAtIndex(1)
	hash := server.levelRedis.GetHash(key)
	length := hash.Count()
	reply = IntegerReply(length)
	return
}

func (server *GoRedisServer) OnHDEL(cmd *Command) (reply *Reply) {
	key := cmd.StringAtIndex(1)
	hash := server.levelRedis.GetHash(key)
	fields := cmd.Args()[2:]
	n := hash.Remove(fields...)
	reply = IntegerReply(n)
	return
}

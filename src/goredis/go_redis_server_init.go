package goredis

/**
初始化绑定
*/
func (server *GoRedisServer) initForKeys() {
	server.On("DEL", func(cmd *Command) (reply *Reply) {
		keys := byteToStrings(cmd.Args[1:])
		count, err := server.OnDEL(cmd, keys...)
		if err != nil {
			reply = ErrorReply(err.Error())
		} else {
			reply = IntegerReply(count)
		}
		return
	})
}

func (server *GoRedisServer) initForStrings() {

	server.On("GET", func(cmd *Command) (reply *Reply) {
		key := cmd.StringAtIndex(1)
		value, err := server.OnGET(cmd, key)
		if err != nil {
			reply = ErrorReply(err.Error())
		} else {
			reply = BulkReply(value)
		}
		return
	})

	server.On("SET", func(cmd *Command) (reply *Reply) {
		key := cmd.StringAtIndex(1)
		value := cmd.StringAtIndex(2)
		err := server.OnSET(cmd, key, value)
		if err != nil {
			reply = ErrorReply(err.Error())
		} else {
			reply = StatusReply("OK")
		}
		return
	})

	server.On("MGET", func(cmd *Command) (reply *Reply) {
		keys := byteToStrings(cmd.Args[1:])
		values, err := server.OnMGET(cmd, keys...)
		if err != nil {
			reply = ErrorReply(err.Error())
		} else {
			reply = MultiBulksReply(values)
		}
		return
	})

	server.On("MSET", func(cmd *Command) (reply *Reply) {
		keyvals := byteToStrings(cmd.Args[1:])
		if len(keyvals)%2 != 0 {
			return ErrorReply("Bad Argument Count")
		}
		err := server.OnMSET(cmd, keyvals...)
		if err != nil {
			reply = ErrorReply(err.Error())
		} else {
			reply = StatusReply("OK")
		}
		return
	})
}

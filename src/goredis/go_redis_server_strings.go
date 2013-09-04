package goredis

func (server *GoRedisServer) initStrings() {

	server.On("GET", func(cmd *Command) (reply *Reply) {
		key := cmd.StringAtIndex(1)
		value := server.OnGET(key)
		reply = BulkReply(value)
		return
	})

	server.On("SET", func(cmd *Command) (reply *Reply) {
		key := cmd.StringAtIndex(1)
		value := cmd.StringAtIndex(2)
		err := server.OnSET(key, value)
		if err != nil {
			reply = ErrorReply(err.Error())
		} else {
			reply = StatusReply("OK")
		}
		return
	})

	server.On("MGET", func(cmd *Command) (reply *Reply) {
		keys := byteToStrings(cmd.Args[1:])
		values := server.OnMGET(keys...)
		reply = MultiBulksReply(values)
		return
	})

	server.On("DEL", func(cmd *Command) (reply *Reply) {
		keys := byteToStrings(cmd.Args[1:])
		count := server.OnDEL(keys...)
		reply = IntegerReply(count)
		return
	})
}

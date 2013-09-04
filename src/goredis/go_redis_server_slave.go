package goredis

func (server *GoRedisServer) initSlave() {
	server.On("GR_SLAVEOF", func(cmd *Command) (reply *Reply) {
		reply = StatusReply(cmd.Session().String())
		return
	})
}

package goredis

func (server *GoRedisServer) init() {
	server.RedisServer.init()
	server.initStrings()
	server.initSlave()
}

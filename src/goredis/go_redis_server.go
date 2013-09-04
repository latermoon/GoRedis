package goredis

type GoRedisServer struct {
	RedisServer
}

func NewGoRedisServer() (server *GoRedisServer) {
	server = &GoRedisServer{}
	server.init()
	return
}

func (server *GoRedisServer) OnGET(key string) (val interface{}) {
	val = "Latermoon"
	return
}

func (server *GoRedisServer) OnSET(key string, val string) (err error) {
	return
}

func (server *GoRedisServer) OnDEL(keys ...string) (count int) {
	count = 1
	return
}

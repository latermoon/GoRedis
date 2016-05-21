package server

import (
	. "github.com/latermoon/GoRedis/redis"
)

// http://redis.io/commands#connection

func (s *GoRedisServer) OnPING(r ReplyWriter, c Command) {
	r.WriteReply(StatusReply("PONG"))
}

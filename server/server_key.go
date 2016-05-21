package server

import (
	. "github.com/latermoon/GoRedis/redis"
)

// http://redis.io/commands#generic

func (s *GoRedisServer) OnDEL(r ReplyWriter, c Command) {

}

func (s *GoRedisServer) OnEXISTS(r ReplyWriter, c Command) {

}

func (s *GoRedisServer) OnKEYS(r ReplyWriter, c Command) {

}

func (s *GoRedisServer) OnTYPE(r ReplyWriter, c Command) {
	elemType := s.db.TypeOf(c[1])
	r.WriteReply(StatusReply(elemType.String()))
}

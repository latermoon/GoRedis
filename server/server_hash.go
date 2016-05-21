package server

import (
	. "github.com/latermoon/GoRedis/redis"
)

// http://redis.io/commands#hash

func (s *GoRedisServer) OnHDEL(r ReplyWriter, c Command) {

}

func (s *GoRedisServer) OnHEXISTS(r ReplyWriter, c Command) {

}

func (s *GoRedisServer) OnHGET(r ReplyWriter, c Command) {
	h := s.db.Hash(c[1])
	val, err := h.Get(c[2])
	if err != nil {
		r.WriteReply(ErrorReply(err.Error()))
		return
	}
	r.WriteReply(BulkReply(val))
}

func (s *GoRedisServer) OnHMGET(r ReplyWriter, c Command) {

}

func (s *GoRedisServer) OnHSET(r ReplyWriter, c Command) {
	h := s.db.Hash(c[1])
	if err := h.Set(c[2], c[3]); err != nil {
		r.WriteReply(ErrorReply(err.Error()))
		return
	}
	r.WriteReply(IntegerReply(1))
}

func (s *GoRedisServer) OnHMSET(r ReplyWriter, c Command) {

}

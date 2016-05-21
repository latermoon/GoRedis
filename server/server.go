package server

import (
	. "github.com/latermoon/GoRedis/redis"
	"github.com/latermoon/GoRedis/rocks"
)

type GoRedisServer struct {
	db *rocks.DB
}

func New(db *rocks.DB) *GoRedisServer {
	return &GoRedisServer{db: db}
}

func (s *GoRedisServer) OnPING(r ReplyWriter, c Command) {
	r.WriteReply(StatusReply("PONG"))
}

package goredis_server

import (
	. "github.com/latermoon/GoRedis/src/goredis"
	"net"
)

type SlaveServer struct {
	conn net.Conn
}

func (s *SlaveServer) Push(cmd *Command) {

}

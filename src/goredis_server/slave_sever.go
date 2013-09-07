package goredis_server

import (
	. "../goredis"
	"net"
)

type SlaveServer struct {
	conn net.Conn
}

func (s *SlaveServer) Push(cmd *Command) {

}

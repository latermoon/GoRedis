package goredis

import (
	"net"
)

type SlaveServer struct {
	conn net.Conn
}

func (s *SlaveServer) Push(cmd *Command) {

}

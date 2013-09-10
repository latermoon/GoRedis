package goredis_server

import (
	. "github.com/latermoon/GoRedis/goredis"
	"net"
)

// 从库会话
type SlaveSession struct {
	conn net.Conn
}

func (s *SlaveSession) Push(cmd *Command) {

}

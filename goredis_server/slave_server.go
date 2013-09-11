package goredis_server

import (
	. "../goredis"
	"net"
)

type SlaveServer struct {
	conn net.Conn
	UID  string
}

func NewSlaveServer(conn net.Conn, uid string) (server *SlaveServer) {
	server = &SlaveServer{}
	server.conn = conn
	server.UID = uid
	return
}

func (s *SlaveServer) Send(bs []byte) (err error) {
	_, err = s.conn.Write(bs)
	return
}

func (s *SlaveServer) SendCommand(cmd *Command) (err error) {
	err = s.Send(cmd.Bytes())
	return
}

func (s *SlaveServer) Connection() net.Conn {
	return s.conn
}

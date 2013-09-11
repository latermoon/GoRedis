package goredis_server

import (
	. "../goredis"
	"net"
)

type SlaveServer struct {
	conn         net.Conn
	linkStatus   LinkStatus
	UID          string
	statusChange chan LinkStatus  // 状态改变
	commandchan  chan interface{} // 顺序处理发送队列
}

func NewSlaveServer(conn net.Conn, uid string) (server *SlaveServer) {
	server = &SlaveServer{}
	server.conn = conn
	server.linkStatus = LinkStatusDown
	server.UID = uid
	server.statusChange = make(chan LinkStatus)
	server.commandchan = make(chan interface{}, 1000000) // 缓冲区
	go server.runloop()
	return
}

func (s *SlaveServer) runloop() {
	for {
		status := <-s.statusChange
		switch status {
		case LinkStatusDown:
		case LinkStatusPending:
		case LinkStatusUp:
		default:
			panic("bad link status")
		}
	}
}

func (s *SlaveServer) Send(bs []byte) (err error) {
	s.commandchan <- bs
	return
}

func (s *SlaveServer) SendCommand(cmd *Command) (err error) {
	err = s.Send(cmd.Bytes())
	return
}

func (s *SlaveServer) SetConnection(conn net.Conn) {

}

func (s *SlaveServer) Connection() net.Conn {
	return s.conn
}

package goredis_server

import (
	. "../goredis"
	"net"
	"sync"
)

// 当前状态
type SLStatus int

const (
	SLStatusConnected = iota
	SLStatusDisconnected
)

type SlaveServer struct {
	conn            net.Conn
	linkStatus      LinkStatus
	status          string
	UID             string
	statusChange    chan LinkStatus  // 状态改变
	commandchan     chan interface{} // 顺序处理发送队列
	chanMutex       *sync.Mutex
	shouldStopWrite bool
}

func NewSlaveServer(conn net.Conn, uid string) (server *SlaveServer) {
	server = &SlaveServer{}
	server.conn = conn
	server.linkStatus = LinkStatusDown
	server.UID = uid
	server.statusChange = make(chan LinkStatus)
	server.commandchan = make(chan interface{}, 1000000) // 缓冲区
	server.chanMutex = &sync.Mutex{}
	server.chanMutex.Lock() // 一开始锁住chan的写操作
	go server.runloop()
	return
}

func (s *SlaveServer) runloop() {
	for {
		s.chanMutex.Lock()
		s.chanMutex.Unlock()
		v := <-s.commandchan
		if s.linkStatus == LinkStatusUp {
			err := s.writeToSlave(v)
			if err != nil {
				s.linkStatus = LinkStatusDown
				s.writeToLocal(v)
			}
		} else if s.linkStatus == LinkStatusDown {
			err := s.writeToLocal(v)
			if err != nil {
				panic("fail to write slave ldb")
			}
		}
	}
}

// 队列中的数据只有两个写入方向，一个是远程从库
func (s *SlaveServer) writeToSlave(v interface{}) (err error) {
	return
}

// 另一个是本地LevelDB
func (s *SlaveServer) writeToLocal(v interface{}) (err error) {
	return
}

// 向远程发送本地更新数据
func (s *SlaveServer) sendLocalChanges() {
	// 停止chan写入
}

func (s *SlaveServer) Send(bs []byte) (err error) {
	s.commandchan <- bs
	return
}

func (s *SlaveServer) SendCommand(cmd *Command) (err error) {
	err = s.Send(cmd.Bytes())
	return
}

func (s *SlaveServer) BindConnection(conn net.Conn) {
	s.conn = conn
	s.chanMutex.Lock()
}

func (s *SlaveServer) Connection() net.Conn {
	return s.conn
}

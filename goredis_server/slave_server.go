package goredis_server

import (
	. "../goredis"
	"fmt"
	"github.com/syndtr/goleveldb/leveldb"
	"github.com/syndtr/goleveldb/leveldb/opt"
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
	UID             string
	useLevelDb      bool
	commandchan     chan interface{} // 顺序处理发送队列
	writeMutex      *sync.Mutex
	shouldStopWrite bool
	// leveldb
	db *leveldb.DB
	ro *opt.ReadOptions
	wo *opt.WriteOptions
}

func NewSlaveServer(uid string) (server *SlaveServer) {
	server = &SlaveServer{}
	server.linkStatus = LinkStatusDown
	server.UID = uid
	server.useLevelDb = len(uid) > 0
	server.commandchan = make(chan interface{}, 100000) // 缓冲区
	server.writeMutex = &sync.Mutex{}
	server.writeMutex.Lock() // 一开始锁住chan的写操作
	// leveldb
	if server.useLevelDb {
		server.ro = &opt.ReadOptions{}
		server.wo = &opt.WriteOptions{}
		dbpath := "/tmp/Slave_" + server.UID + ".ldb"
		server.db, err = leveldb.OpenFile(dbpath, &opt.Options{Flag: opt.OFCreateIfMissing})
		if err != nil {
			fmt.Println("slave db error", err)
			server.useLevelDb = false
		}
	}
	return
}

func (s *SlaveServer) runloop() {
	for {
		s.writeMutex.Lock()
		s.writeMutex.Unlock()
		v := <-s.commandchan
		if s.linkStatus == LinkStatusUp {
			err := s.writeToSlave(v)
			if err != nil {
				s.linkStatus = LinkStatusDown
				if s.useLevelDb {
					s.writeToLocal(v)
				} else {
					// 抛错，中断Slave Connection，撤销实例
				}
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
func (s *SlaveServer) sendLocalChanges() (err error) {
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
	// 绑定连接的时候，如果有绑定本地数据库，就发送出去
	if s.useLevelDb {
		s.sendLocalChanges()
	}
	// 发送完毕后开始消化
	go server.runloop()
}

func (s *SlaveServer) Connection() net.Conn {
	return s.conn
}

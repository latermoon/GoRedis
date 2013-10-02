package goredis_server

import (
	. "../goredis"
	"fmt"
	"github.com/syndtr/goleveldb/leveldb"
	"github.com/syndtr/goleveldb/leveldb/opt"
	"sync"
)

/*
new slave
& do nothing

if ldb: set conn: nil
	writeToldb yes
set conn: net.Conn
	wait stopWrite
	send ldbq
		err:continue writeToldb
	writeToSlave
		err:continue writeToldb

*/
type SlaveServer struct {
	session    *Session
	linkStatus LinkStatus
	UID        string

	commandchan     chan interface{} // 顺序处理发送队列
	chan1           chan int
	writeMutex      *sync.Mutex
	shouldStopWrite bool

	useLevelDb bool
	// leveldb
	db *leveldb.DB
	ro *opt.ReadOptions
	wo *opt.WriteOptions
}

func NewSlaveServer(uid string) (server *SlaveServer) {
	server = &SlaveServer{}
	server.linkStatus = LinkStatusInit
	server.UID = uid
	server.useLevelDb = len(uid) > 0
	server.commandchan = make(chan interface{}, 1000) // 缓冲区
	server.chan1 = make(chan int)
	server.writeMutex = &sync.Mutex{}
	// leveldb
	if server.useLevelDb {
		server.ro = &opt.ReadOptions{}
		server.wo = &opt.WriteOptions{}
		dbpath := "/tmp/Slave_" + server.UID + ".ldb"
		var e1 error
		server.db, e1 = leveldb.OpenFile(dbpath, &opt.Options{Flag: opt.OFCreateIfMissing})
		if e1 != nil {
			fmt.Println("slave db error", e1)
			server.useLevelDb = false
		}
	}
	return
}

func (s *SlaveServer) runloop() {
	for {
		if s.linkStatus == LinkStatusPending {
			fmt.Println("Pending exit runloop")
			return
		}

		if s.shouldStopWrite {
			fmt.Println("shouldStopWrite")
			s.shouldStopWrite = false
			return
		}

		v := <-s.commandchan
		switch s.linkStatus {
		case LinkStatusUp:
			err := s.writeToSlave(v)
			if err != nil {
				s.db.Close()
				s.db = nil
				s.linkStatus = LinkStatusDown
				if s.useLevelDb {
					s.writeToLocal(v)
				} else {
					// 抛错，中断Slave Connection，撤销实例
				}
			}
		case LinkStatusDown:
			err := s.writeToLocal(v)
			if err != nil {
				panic("fail to write slave ldb")
			}
		}

	}
}

// 队列中的数据只有两个写入方向，一个是远程从库
func (s *SlaveServer) writeToSlave(v interface{}) (err error) {
	fmt.Println("writeToSlave...")
	_, err = s.session.Write(v.([]byte))
	return
}

// 另一个是本地LevelDB
func (s *SlaveServer) writeToLocal(v interface{}) (err error) {
	fmt.Println("writeToLocal...")
	return
}

// 向远程发送本地更新数据
func (s *SlaveServer) sendLocalChanges() (err error) {
	fmt.Println("sendLocalChanges")
	return
}

func (s *SlaveServer) SendBytes(bs []byte) (err error) {
	s.commandchan <- bs
	return
}

func (s *SlaveServer) SendCommand(cmd *Command) (err error) {
	err = s.SendBytes(cmd.Bytes())
	return
}

func (s *SlaveServer) SetSession(session *Session) {
	s.session = session
}

func (s *SlaveServer) Active() {
	go s.runloop()
}

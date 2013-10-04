package goredis_server

import (
	. "../goredis"
	// . "./storage"
	"fmt"
	"github.com/syndtr/goleveldb/leveldb"
	"github.com/syndtr/goleveldb/leveldb/opt"
	"sync"
)

// 指向从库的会话
type SlaveSession struct {
	session   *Session
	cmdbuffer chan *Command
	sendmutex sync.Mutex
}

func NewSlaveSession(session *Session) (s *SlaveSession) {
	s = &SlaveSession{}
	s.session = session
	s.cmdbuffer = make(chan *Command, 100000)
	return
}

func (s *SlaveSession) runloop() {
	for {
		// s.sendmutex.Lock()
		// defer s.sendmutex.Unlock()
		cmd := <-s.cmdbuffer
		err := s.session.WriteCommand(cmd)
		if err != nil {
			panic("bad connecton...")
		}
	}
}

func (s *SlaveSession) SendCommand(cmd *Command) {
	s.cmdbuffer <- cmd
}

// 向从库发送数据库快照
func (s *SlaveSession) SendSnapshot(snapshot *leveldb.Snapshot) {
	s.sendmutex.Lock()
	defer s.sendmutex.Unlock()
	iter := snapshot.NewIterator(&opt.ReadOptions{})
	for iter.Next() {
		fmt.Println(string(iter.Key()), string(iter.Value()))
	}
	iter.Release()
	snapshot.Release()
	go s.runloop()
}

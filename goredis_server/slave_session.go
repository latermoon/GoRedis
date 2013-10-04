package goredis_server

import (
	. "../goredis"
	. "./storage"
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
		cmd := <-s.cmdbuffer
		fmt.Println("send", cmd)
		err := s.session.WriteCommand(cmd)
		if err != nil {
			fmt.Println("slave gone away ...")
			return
		}
	}
}

func (s *SlaveSession) SendCommand(cmd *Command) {
	s.cmdbuffer <- cmd
}

// 向从库发送数据库快照
// 时间关系，怎是使用了 []byte -> Entry -> Command -> slave 的方法，
// 应该改为官方发送rdb数据的方式
func (s *SlaveSession) SendSnapshot(snapshot *leveldb.Snapshot) {
	s.sendmutex.Lock()
	defer s.sendmutex.Unlock()

	iter := snapshot.NewIterator(&opt.ReadOptions{})
	for iter.Next() {
		entry, e1 := s.toEntry(iter.Value())
		if e1 != nil {
			fmt.Println(e1)
			continue
		}
		cmd := entryToCommand(iter.Key(), entry)
		fmt.Println("cmd", cmd)
		if cmd != nil {
			s.session.WriteCommand(cmd)
		} else {
			fmt.Println(string(iter.Key()), string(iter.Value()))
		}
	}
	// 必须释放Iterator和Snapshot
	iter.Release()
	snapshot.Release()
	// 开始消费
	go s.runloop()
}

func (s *SlaveSession) toEntry(bs []byte) (entry Entry, err error) {
	switch EntryType(bs[0]) {
	case EntryTypeString:
		entry = NewStringEntry(nil)
	case EntryTypeHash:
		entry = NewHashEntry()
	case EntryTypeSortedSet:
		entry = NewSortedSetEntry()
	case EntryTypeSet:
		entry = NewSetEntry()
	case EntryTypeList:
		entry = NewListEntry()
	default:
		entry = NewStringEntry(nil)
	}
	// 反序列化
	err = entry.Decode(bs[1:])
	if err != nil {
		fmt.Println("decode entry error", err)
	}
	return
}

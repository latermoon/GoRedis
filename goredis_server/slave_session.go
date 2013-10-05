package goredis_server

import (
	. "../goredis"
	"./libs/leveltool"
	. "./storage"
	"fmt"
	"github.com/syndtr/goleveldb/leveldb"
	"github.com/syndtr/goleveldb/leveldb/opt"
	"strings"
	"sync"
)

// 指向从库的会话
// new slave: sendSnapshot -> remoteRunloop -> aofRunloop -> aofToRemoteRunloop -> remoteRunloop
// old slave: aofRunloop -> aofToRemoteRunloop -> remoteRunloop
type SlaveSession struct {
	session              *Session
	server               *GoRedisServer
	cmdbuffer            chan *Command
	currentCommand       *Command // 当前处理中的command
	sendmutex            sync.Mutex
	uid                  string               // 从库id
	aofEnabled           bool                 // 支持从库快照
	remoteExists         bool                 //是否存在远程连接
	shouldChangeToRemote bool                 // 应该转向写入远程
	aoflist              *leveltool.LevelList // 存放同步中断后指令
}

func NewSlaveSession(server *GoRedisServer, session *Session, uid string) (s *SlaveSession) {
	s = &SlaveSession{}
	s.server = server
	s.session = session
	s.remoteExists = s.session != nil
	s.cmdbuffer = make(chan *Command, 100000)
	s.uid = uid
	s.aofEnabled = len(uid) > 0 //存在uid就可以访问本地快照
	if s.aofEnabled {
		// TODO aof之后存到另外的路径，节省主库空间
		s.aoflist = leveltool.NewLevelList(s.server.datasource.(*LevelDBDataSource).DB(), "__goredis:slaveaof:"+s.uid)
	}
	return
}

func (s *SlaveSession) AofEnabled() bool {
	return s.aofEnabled
}

func (s *SlaveSession) UID() string {
	return s.uid
}

func (s *SlaveSession) SetSession(session *Session) {
	s.session = session
	s.remoteExists = s.session != nil
}

// 继续同步
func (s *SlaveSession) ContinueSync() {
	s.shouldChangeToRemote = true
}

func (s *SlaveSession) ContinueAof() {
	go s.aofRunloop()
}

// 向远程写入
func (s *SlaveSession) remoteRunloop() {
	for {
		// 先消费别人的
		if s.currentCommand == nil {
			s.currentCommand = <-s.cmdbuffer
		}
		fmt.Println("send", s.currentCommand)
		err := s.session.WriteCommand(s.currentCommand)
		if err != nil {
			fmt.Println("slave gone away ...")
			// 从库断开后写入本地aof
			if s.AofEnabled() {
				fmt.Println("change to aof ...")
				go s.aofRunloop()
				return
			}
			return
		}
		s.currentCommand = <-s.cmdbuffer
	}
}

// 向本地aof写入
func (s *SlaveSession) aofRunloop() {
	for {
		if s.currentCommand == nil {
			s.currentCommand = <-s.cmdbuffer
		}
		fmt.Println("aof push", s.currentCommand)
		_, err := s.aoflist.Push(s.currentCommand.Bytes())
		// 如果写入aof出错，应该废弃全部aof，重来snapshot
		if err != nil {
			fmt.Println("Error aof push ...", err)
			return
		}
		if s.shouldChangeToRemote {
			s.currentCommand = nil // 清空并跳转
			s.shouldChangeToRemote = false
			go s.aofToRemoteRunloop()
			return
		}
		s.currentCommand = <-s.cmdbuffer
	}
}

// 从aof读取向远程写入
func (s *SlaveSession) aofToRemoteRunloop() {
	// 从aof向远程写时，不应该有待处理的数据
	if s.currentCommand != nil {
		fmt.Println("where are you?", s.currentCommand)
		return
	}
	sendCount := 0
	for {
		elem, e1 := s.aoflist.Index(0)
		if e1 != nil {
			// 如果aof出错，应该废弃全部aof，重来snapshot
			fmt.Println("Error aof index ...", e1)
			return
		}
		// 同步完毕，转向直接远程写入
		if elem == nil {
			fmt.Println("aof finish count", sendCount)
			go s.remoteRunloop()
			return
		}
		sendCount++
		// bs come from cmd.Bytes()
		bs := elem.Value.([]byte)
		n, e2 := s.session.Write(bs)
		if e2 != nil {
			fmt.Println("aof to remote error", n, e2)
			return
		}
		// 移除
		_, e3 := s.aoflist.Pop()
		if e3 != nil {
			// 如果aof出错，应该废弃全部aof，重来snapshot
			fmt.Println("Error aof pop ...", e3)
			return
		}
	}
}

func (s *SlaveSession) AsyncSendCommand(cmd *Command) {
	s.cmdbuffer <- cmd
}

// 向从库发送数据库快照
// 时间关系，暂时使用了 []byte -> Entry -> Command -> slave 的方法，
// 应该改为官方发送rdb数据的方式
func (s *SlaveSession) SendSnapshot(snapshot *leveldb.Snapshot) {
	s.sendmutex.Lock()
	defer s.sendmutex.Unlock()

	iter := snapshot.NewIterator(&opt.ReadOptions{})
	defer func() {
		// 必须释放Iterator和Snapshot
		iter.Release()
		snapshot.Release()
	}()

	for iter.Next() {
		// 跳过系统数据
		key := string(iter.Key())
		if strings.HasPrefix(key, "__goredis:") {
			continue
		}
		entry, e1 := s.toEntry(iter.Value())
		if e1 != nil {
			fmt.Println(e1)
			continue
		}
		cmd := entryToCommand(iter.Key(), entry)
		if cmd == nil {
			fmt.Println(string(iter.Key()), string(iter.Value()))
			continue
		}
		fmt.Println("cmd", cmd)
		e2 := s.session.WriteCommand(cmd)
		if e2 != nil {
			// 销毁整个slave
		}
	}
	// 构建aof
	if s.AofEnabled() {
		s.server.snapshotSentCallback(s)
	}
	// 开始消费
	go s.remoteRunloop()
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

package goredis_server

import (
	. "../goredis"
	// "./libs/levelredis"
	qp "./libs/queueprocess"
	"strings"
)

type keyValuePair struct {
	Key       interface{}
	Value     interface{}
	EntryType EntryType
}

// 主从同步中的从库连接
type SlaveSessionClient struct {
	session           *Session
	server            *GoRedisServer
	taskqueue         *qp.QueueProcess // 队列处理
	shouldStopRunloop bool             // 跳出runloop指令
}

func NewSlaveSessionClient(server *GoRedisServer, session *Session) (s *SlaveSessionClient) {
	s = &SlaveSessionClient{}
	s.server = server
	s.session = session
	s.taskqueue = qp.NewQueueProcess(100, s.queueHandler)
	return
}

// 获取redis info
func (s *SlaveSessionClient) detectRedisInfo(section string) (info string, err error) {
	cmdinfo := NewCommand([]byte("INFO"), []byte(section))
	s.session.WriteCommand(cmdinfo)
	var reply *Reply
	reply, err = s.session.ReadReply()
	if err == nil {
		switch reply.Value.(type) {
		case string:
			info = reply.Value.(string)
		case []byte:
			info = string(reply.Value.([]byte))
		default:
			info = reply.String()
		}
	}
	return
}

func (s *SlaveSessionClient) Start() {
	if s.shouldStopRunloop {
		s.server.stdlog.Error("[%s] slaveof should run once", s.session.RemoteAddr())
		return
	}
	// 异步入库
	// 阻塞处理，直到出错
	err := s.startSync()
	if err != nil {
		s.server.stdlog.Error("[%s] slaveof sync error %s", s.session.RemoteAddr(), err)
	}
	// 终止运行
	s.shouldStopRunloop = true
}

func (s *SlaveSessionClient) queueHandler(t qp.Task) {
	if s.shouldStopRunloop {
		s.taskqueue.Stop()
		s.server.stdlog.Info("[%s] slaveof stop runloop", s.session.RemoteAddr())
		return
	}
	obj := t
	// Process
	s.server.syncCounters.Get("buffer").SetCount(s.taskqueue.Len())
	// s.server.stdlog.Debug("[%s] slaveof process %s", s.session.RemoteAddr(), obj)
	switch obj.(type) {
	case *Command:
		// 这些sync回来的command，全部是更新操作，不需要返回reply
		cmd := obj.(*Command)
		_ = s.server.InvokeCommandHandler(s.session, cmd)
		s.server.syncCounters.Get("total").Incr(1)
		cmdName := strings.ToUpper(cmd.Name())
		cate := GetCommandCategory(cmdName)
		// incr counter
		s.server.syncCounters.Get(string(cate)).Incr(1)
		switch cmdName {
		case "PING":
			s.server.syncCounters.Get("ping").Incr(1)
		}
	case *keyValuePair:
		kv := obj.(*keyValuePair)
		entryKey := string(kv.Key.([]byte))
		s.server.syncCounters.Get("total").Incr(1)
		switch kv.EntryType {
		case EntryTypeString:
			s.server.syncCounters.Get("string").Incr(1)
			s.server.levelRedis.Strings().Set(kv.Key.([]byte), kv.Value.([]byte))
		case EntryTypeHash:
			s.server.syncCounters.Get("hash").Incr(1)
			s.server.levelRedis.GetHash(entryKey).Set(kv.Value.([][]byte)...)
		case EntryTypeList:
			s.server.syncCounters.Get("list").Incr(1)
			s.server.levelRedis.GetList(entryKey).RPush(kv.Value.([][]byte)...)
		case EntryTypeSet:
			s.server.syncCounters.Get("set").Incr(1)
			s.server.levelRedis.GetSet(entryKey).Set(kv.Value.([][]byte)...)
		case EntryTypeSortedSet:
			s.server.syncCounters.Get("zset").Incr(1)
			zset := s.server.levelRedis.GetSortedSet(entryKey)
			zset.Add(kv.Value.([][]byte)...)
		default:
			s.server.stdlog.Warn("[%s] bad entry type", s.session.RemoteAddr())
		}
	default:
		s.server.stdlog.Warn("[%s] bad queue obj", s.session.RemoteAddr())
	}
}

func (s *SlaveSessionClient) startSync() (err error) {
	// 检查是goredis还是官方redis
	var info string
	info, err = s.detectRedisInfo("server")
	if err != nil {
		s.server.stdlog.Error("[%s] slave of error %s", s.session.RemoteAddr(), err)
		return
	}
	isGoRedis := strings.Index(info, "goredis_version") > 0

	var cmdsync *Command
	if isGoRedis {
		// 如果是GoRedis，需要发送自身uid，实现增量同步
		cmdsync = NewCommand([]byte("SYNC"), []byte("uid"), []byte(s.server.UID()))
	} else {
		// 官方Redis的SYNC不接受多参数，-ERR wrong number of arguments for 'sync' command
		cmdsync = NewCommand([]byte("SYNC"))
	}

	s.session.WriteCommand(cmdsync)

	// 这里代码有有点乱，可优化
	readRdbFinish := false
	var c byte
	for {
		c, err = s.session.PeekByte()
		if err != nil {
			s.server.stdlog.Warn("master gone away %s", s.session.RemoteAddr())
			break
		}
		if c == '*' {
			if cmd, e2 := s.session.ReadCommand(); e2 == nil {
				// PUSH
				hash := 0
				if len(cmd.Args) > 1 {
					hash = qp.StringCharSum(cmd.StringAtIndex(1))
				}
				s.taskqueue.Process(hash, cmd)
			} else {
				s.server.stdlog.Error("sync error %s", e2)
				err = e2
				break
			}
		} else if !readRdbFinish && c == '$' {
			s.server.stdlog.Info("[%s] sync rdb ", s.session.RemoteAddr())
			s.session.ReadByte()
			var rdbsize int
			rdbsize, err = s.session.ReadLineInteger()
			if err != nil {
				break
			}
			s.server.stdlog.Info("[%s] rdb size %d bytes", s.session.RemoteAddr(), rdbsize)
			// read
			// dec := newDecoder(s.server, s)
			// err = rdb.Decode(s.session, dec)
			// if err != nil {
			// 	break
			// }
			readRdbFinish = true
		} else {
			s.server.stdlog.Debug("[%s] skip byte %q %s", s.session.RemoteAddr(), c, string(c))
			_, err = s.session.ReadByte()
			if err != nil {
				break
			}
		}
	}
	return
}

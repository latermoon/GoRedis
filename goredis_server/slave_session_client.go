package goredis_server

import (
	. "../goredis"
	"./libs/rdb"
	. "./storage"
	"strings"
	"time"
)

type keyValuePair struct {
	Key   interface{}
	Value interface{}
}

// 主从同步中的从库连接
type SlaveSessionClient struct {
	session           *Session
	server            *GoRedisServer
	taskqueue         *SafeList // 队列存储
	shouldStopRunloop bool      // 跳出runloop指令
}

func NewSlaveSessionClient(server *GoRedisServer, session *Session) (s *SlaveSessionClient) {
	s = &SlaveSessionClient{}
	s.server = server
	s.session = session
	s.taskqueue = NewSafeList()
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
	go s.processRunloop()
	// 阻塞处理，直到出错
	err := s.startSync()
	if err != nil {
		s.server.stdlog.Error("[%s] slaveof sync error %s", s.session.RemoteAddr(), err)
	}
	// 终止运行
	s.shouldStopRunloop = true
}

func (s *SlaveSessionClient) processRunloop() {
	var sleepCount uint
	sleepCount = 0
	for {
		if s.shouldStopRunloop {
			s.server.stdlog.Info("[%s] slaveof stop runloop", s.session.RemoteAddr())
			break
		}
		// POP
		obj := s.taskqueue.LPop()
		if obj == nil {
			sleepCount++
			if sleepCount >= 100 || sleepCount <= 0 {
				s.server.stdlog.Info("[%s] slaveof still waiting for data ", s.session.RemoteAddr())
				sleepCount = 1
			}
			sleepTime := time.Millisecond * time.Duration(10*sleepCount)
			time.Sleep(sleepTime)
			continue
		} else {
			sleepCount = 0
		}
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
			switch cmdName {
			case "SET":
				s.server.syncCounters.Get("string").Incr(1)
			case "HSET":
				s.server.syncCounters.Get("hash").Incr(1)
			case "SADD":
				s.server.syncCounters.Get("set").Incr(1)
			case "RPUSH":
				s.server.syncCounters.Get("list").Incr(1)
			case "ZADD":
				s.server.syncCounters.Get("zset").Incr(1)
			case "PING":
				s.server.syncCounters.Get("ping").Incr(1)
			}
		case *keyValuePair:
			key := obj.(*keyValuePair).Key.([]byte)
			entry := obj.(*keyValuePair).Value.(Entry)
			s.server.syncCounters.Get("total").Incr(1)
			switch entry.Type() {
			case EntryTypeString:
				s.server.syncCounters.Get("string").Incr(1)
			case EntryTypeHash:
				s.server.syncCounters.Get("hash").Incr(1)
			case EntryTypeList:
				s.server.syncCounters.Get("list").Incr(1)
			case EntryTypeSet:
				s.server.syncCounters.Get("set").Incr(1)
			case EntryTypeSortedSet:
				s.server.syncCounters.Get("zset").Incr(1)
			default:
				s.server.stdlog.Warn("[%s] bad entry type", s.session.RemoteAddr())
			}
			e2 := s.server.datasource.Set(string(key), entry)
			if e2 != nil {
				s.server.stdlog.Error("[%s] datasource set error %s", s.session.RemoteAddr(), e2)
			}
		default:
			s.server.stdlog.Warn("[%s] bad queue obj", s.session.RemoteAddr())
		}
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
				s.taskqueue.RPush(cmd)
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
			dec := newDecoder(s.server, s)
			err = rdb.Decode(s.session, dec)
			if err != nil {
				break
			}
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

// 第三方rdb解释函数
type rdbDecoder struct {
	db       int
	i        int
	keyCount int
	rdb.NopDecoder
	server      *GoRedisServer
	slaveClient *SlaveSessionClient
	// 数据缓冲区
	stringEntry *StringEntry
	hashEntry   *HashEntry
	setEntry    *SetEntry
	listEntry   *ListEntry
	zsetEntry   *SortedSetEntry
}

func newDecoder(server *GoRedisServer, slaveClient *SlaveSessionClient) (dec *rdbDecoder) {
	dec = &rdbDecoder{}
	dec.server = server
	dec.slaveClient = slaveClient
	dec.keyCount = 0
	return
}

func (p *rdbDecoder) StartDatabase(n int) {
	p.db = n
}

func (p *rdbDecoder) EndDatabase(n int) {

}

func (p *rdbDecoder) EndRDB() {
	p.server.stdlog.Info("[%s] rdb end, sync %d items", p.slaveClient.session.RemoteAddr(), p.keyCount)
}

// Set
func (p *rdbDecoder) Set(key, value []byte, expiry int64) {
	p.keyCount++
	p.stringEntry = NewStringEntry(string(value))
	p.slaveClient.taskqueue.RPush(&keyValuePair{Key: key, Value: p.stringEntry})
}

func (p *rdbDecoder) StartHash(key []byte, length, expiry int64) {
	p.keyCount++
	p.hashEntry = NewHashEntry()
}

func (p *rdbDecoder) Hset(key, field, value []byte) {
	p.hashEntry.Set(string(field), string(value))
}

// Hash
func (p *rdbDecoder) EndHash(key []byte) {
	p.slaveClient.taskqueue.RPush(&keyValuePair{Key: key, Value: p.hashEntry})
}

func (p *rdbDecoder) StartSet(key []byte, cardinality, expiry int64) {
	p.keyCount++
	p.setEntry = NewSetEntry()
}

func (p *rdbDecoder) Sadd(key, member []byte) {
	p.setEntry.Put(string(member))
}

// Set
func (p *rdbDecoder) EndSet(key []byte) {
	p.slaveClient.taskqueue.RPush(&keyValuePair{Key: key, Value: p.setEntry})
}

func (p *rdbDecoder) StartList(key []byte, length, expiry int64) {
	p.keyCount++
	p.listEntry = NewListEntry()
	p.i = 0
}

func (p *rdbDecoder) Rpush(key, value []byte) {
	p.listEntry.List().RPush(string(value))
	p.i++
}

// List
func (p *rdbDecoder) EndList(key []byte) {
	p.slaveClient.taskqueue.RPush(&keyValuePair{Key: key, Value: p.listEntry})
}

func (p *rdbDecoder) StartZSet(key []byte, cardinality, expiry int64) {
	p.keyCount++
	p.zsetEntry = NewSortedSetEntry()
	p.i = 0
}

func (p *rdbDecoder) Zadd(key []byte, score float64, member []byte) {
	p.zsetEntry.SortedSet().Add(string(member), score)
	p.i++
}

// ZSet
func (p *rdbDecoder) EndZSet(key []byte) {
	p.slaveClient.taskqueue.RPush(&keyValuePair{Key: key, Value: p.zsetEntry})
}

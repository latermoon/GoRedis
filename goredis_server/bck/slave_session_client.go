package goredis_server

/*
slaveClient.Start()
slaveClient.Stop()
*/
import (
	"./monitor"
	. "GoRedis/goredis"
	"GoRedis/libs/levelredis"
	qp "GoRedis/libs/queueprocess"
	"GoRedis/libs/stdlog"
	"errors"
	"fmt"
	// "github.com/latermoon/levigo"
	levigo "github.com/bsm/go-rocksdb"
	"github.com/latermoon/msgpackgo/codec"
	"os"
	"time"
)

// 主从同步中的从库连接
type SlaveSessionClient struct {
	session           *SlaveSession
	server            *GoRedisServer
	taskqueue         *qp.QueueProcess // 队列处理
	shouldStopRunloop bool             // 跳出runloop指令
	aofRedis          *levelredis.LevelRedis
	queueCount        int // 队列数
	queueLists        []*levelredis.LevelList
	msgpack_handle    codec.MsgpackHandle
	// 监控
	syncCounters *monitor.Counters
	syncMonitor  *monitor.StatusLogger
	// path
	homedir string
}

func NewSlaveSessionClient(server *GoRedisServer, session *SlaveSession) (s *SlaveSessionClient) {
	s = &SlaveSessionClient{}
	s.server = server
	s.session = session
	// s.taskqueue = qp.NewQueueProcess(100, s.queueHandler)
	s.queueCount = 100
	s.homedir = fmt.Sprintf("%sslaveof_%s", s.server.directory, s.session.RemoteAddr())
	os.MkdirAll(s.homedir, os.ModePerm)
	s.initMonitor()
	s.initSlaveDb()
	return
}

func (s *SlaveSessionClient) initSlaveDb() (err error) {
	opts := levigo.NewOptions()
	opts.SetCache(levigo.NewLRUCache(32 * 1024 * 1024))
	opts.SetCompression(levigo.SnappyCompression)
	opts.SetBlockSize(32 * 1024)
	opts.SetWriteBufferSize(128 * 1024 * 1024)
	opts.SetCreateIfMissing(true)
	db, e1 := levigo.Open(s.homedir+"/db0", opts)
	if e1 != nil {
		return e1
	}
	s.aofRedis = levelredis.NewLevelRedis(db)
	// init lists
	s.queueLists = make([]*levelredis.LevelList, s.queueCount)
	for i := 0; i < s.queueCount; i++ {
		aofkey := fmt.Sprintf("queue_%d", i)
		s.queueLists[i] = s.aofRedis.GetList(aofkey)
	}
	return
}

func (s *SlaveSessionClient) initMonitor() {
	s.syncCounters = monitor.NewCounters()
	s.syncMonitor = monitor.NewStatusLogger(s.homedir + "/sync.log")
	s.syncMonitor.Add(monitor.NewTimeFormater("Time", 8))
	cmds := []string{"rdbsync", "cmdsync", "proc"}
	for _, cmd := range cmds {
		s.syncMonitor.Add(monitor.NewCountFormater(s.syncCounters.Get(cmd), cmd, 8, "ChangedCount"))
	}
	// buffer用于显示同步过程中的taskqueue buffer长度
	s.syncMonitor.Add(monitor.NewCountFormater(s.syncCounters.Get("buffer"), "buffer", 9, "Count"))
	go s.syncMonitor.Start()
}

func (s *SlaveSessionClient) Start() {
	if s.shouldStopRunloop {
		stdlog.Printf("[%s] slaveof should run once\n", s.session.RemoteAddr())
		return
	}
	// 阻塞处理，直到出错
	s.session.DidRecvCommand = s.didRecvCommand
	s.session.RdbFinished = s.rdbFinished
	err := s.session.Sync(s.server.UID())
	if err != nil {
		stdlog.Printf("[%s] slaveof sync error %s\n", s.session.RemoteAddr(), err)
	}
	// 终止运行
	s.shouldStopRunloop = true
}

func (s *SlaveSessionClient) didRecvCommand(cmd *Command, count int64, isrdb bool) {
	if len(cmd.Args) == 1 {
		return
	}
	s.syncCounters.Get("buffer").Incr(1)
	if isrdb {
		s.syncCounters.Get("rdbsync").Incr(1)
	} else {
		s.syncCounters.Get("cmdsync").Incr(1)
	}
	key, _ := cmd.ArgAtIndex(1)
	lst := s.queueLists[SumOfBytesChars(key)%s.queueCount]
	out, err := s.encodeCommand(cmd)
	if err == nil {
		lst.RPush(out)
	} else {
		fmt.Println("err,", err)
	}
}

// 当rdb同步结束后，开始启动消费队列
func (s *SlaveSessionClient) rdbFinished(count int64) {
	for i := 0; i < s.queueCount; i++ {
		go s.queueProcess(i)
	}
}

func (s *SlaveSessionClient) queueProcess(i int) {
	lst := s.queueLists[i]
	for {
		if s.shouldStopRunloop {
			return
		}
		if lst.Len() == 0 {
			time.Sleep(time.Millisecond * time.Duration(100))
		}
		elem, e1 := lst.LPop()
		if e1 != nil {
			fmt.Println("lpop err", i, e1)
			time.Sleep(time.Millisecond * time.Duration(100))
			continue
		}
		if elem == nil {
			continue
		}
		s.syncCounters.Get("buffer").Incr(-1)
		s.syncCounters.Get("proc").Incr(1)
		cmd, e2 := s.decodeCommand(elem.Value.([]byte))
		if e2 != nil {
			fmt.Println("decode err", i, e1)
			time.Sleep(time.Millisecond * time.Duration(100))
			continue
		}
		_ = s.server.InvokeCommandHandler(s.session.Session(), cmd)
		// fmt.Println(aofkey, lst.Len(), "->pop", cmd)
		// cmd := NewCommand(obj.([]byte)...)
	}
}

func (s *SlaveSessionClient) encodeCommand(cmd *Command) (out []byte, err error) {
	enc := codec.NewEncoderBytes(&out, &s.msgpack_handle)
	err = enc.Encode(cmd.Args)
	return
}

func (s *SlaveSessionClient) decodeCommand(in []byte) (cmd *Command, err error) {
	dec := codec.NewDecoderBytes(in, &s.msgpack_handle)
	var v interface{}
	err = dec.Decode(&v)
	if err == nil {
		objs, ok := v.([]interface{})
		if !ok {
			err = errors.New("bad command bytes")
			return
		}
		args := make([][]byte, 0, len(objs))
		for i := 0; i < len(objs); i++ {
			args = append(args, objs[i].([]byte))
		}
		cmd = NewCommand(args...)
	}
	return
}

func (s *SlaveSessionClient) queueHandler(t qp.Task) {
}

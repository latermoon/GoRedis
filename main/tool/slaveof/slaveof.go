package slaveof

/*

*/
import (
	. "../../../goredis"
	. "../../../goredis_server/"
	"../../../goredis_server/monitor"
	"../../../libs/levelredis"
	"../../../libs/safelist"
	"errors"
	"fmt"
	"github.com/latermoon/levigo"
	"github.com/latermoon/msgpackgo/codec"
	"github.com/latermoon/redigo/redis"
	"net"
	"os"
	"time"
)

// 主从同步中的从库连接
type SlaveOf struct {
	session           *SlaveSession
	shouldStopRunloop bool // 跳出runloop指令
	aofRedis          *levelredis.LevelRedis
	queueCount        int // 队列数
	queueLists        []*safelist.SafeList
	msgpack_handle    codec.MsgpackHandle
	// 监控
	syncCounters *monitor.Counters
	syncMonitor  *monitor.StatusLogger
	// path
	homedir string
	// hosts
	srchost string
	dsthost string
	// pool
	pool *redis.Pool
}

func NewSlaveOf(homedir string, srchost, dsthost string) (s *SlaveOf, err error) {
	s = &SlaveOf{}
	s.srchost = srchost
	s.dsthost = dsthost
	s.queueCount = 100
	err = s.initConnection()
	if err != nil {
		return
	}
	s.homedir = homedir
	os.MkdirAll(s.homedir, os.ModePerm)
	err = s.initSlaveDb()
	if err != nil {
		return
	}
	s.initMonitor()
	return
}

func (s *SlaveOf) initConnection() (err error) {
	srcconn, e1 := net.Dial("tcp", s.srchost)
	if e1 != nil {
		return e1
	}
	s.session = NewSlaveSession(NewSession(srcconn), s.srchost)
	s.pool = &redis.Pool{
		MaxIdle:     500,
		IdleTimeout: 240 * time.Second,
		Dial: func() (redis.Conn, error) {
			c, err := redis.Dial("tcp", s.dsthost)
			return c, err
		},
	}
	return
}

func (s *SlaveOf) initSlaveDb() (err error) {
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
	s.queueLists = make([]*safelist.SafeList, s.queueCount)
	for i := 0; i < s.queueCount; i++ {
		// aofkey := fmt.Sprintf("queue_%d", i)
		// s.queueLists[i] = s.aofRedis.GetList(aofkey)
		s.queueLists[i] = safelist.NewSafeList()
	}
	return
}

func (s *SlaveOf) initMonitor() {
	s.syncCounters = monitor.NewCounters()
	s.syncMonitor = monitor.NewStatusLogger(s.homedir + "/sync.log")
	s.syncMonitor.Add(monitor.NewTimeFormater("Time", 8))
	cmds := []string{"rdbsync", "cmdsync", "proc"}
	for _, cmd := range cmds {
		s.syncMonitor.Add(monitor.NewCountFormater(s.syncCounters.Get(cmd), cmd, 16, "ChangedCount"))
	}
	// buffer用于显示同步过程中的taskqueue buffer长度
	s.syncMonitor.Add(monitor.NewCountFormater(s.syncCounters.Get("buffer"), "buffer", 16, "Count"))
	go s.syncMonitor.Start()
}

func (s *SlaveOf) Start() {
	if s.shouldStopRunloop {
		fmt.Printf("[%s] slaveof should run once\n", s.session.RemoteAddr())
		return
	}
	// 阻塞处理，直到出错
	s.session.DidRecvCommand = s.didRecvCommand
	s.session.RdbFinished = s.rdbFinished
	err := s.session.Sync("")
	if err != nil {
		fmt.Printf("[%s] slaveof sync error %s\n", s.session.RemoteAddr(), err)
	}
	// 终止运行
	s.shouldStopRunloop = true
}

func (s *SlaveOf) didRecvCommand(cmd *Command, count int64, isrdb bool) {
	if len(cmd.Args) == 1 {
		return
	}
	// skip
	if cmd.StringAtIndex(1) == "user:update:timestamp" {
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
	// out, err := s.encodeCommand(cmd)
	// if err == nil {
	// 	lst.RPush(out)
	// } else {
	// 	fmt.Println("err,", err)
	// }
	lst.RPush(cmd)
	if count%100 == 0 {
		time.Sleep(time.Millisecond * 1)
	}
}

// 当rdb同步结束后，开始启动消费队列
func (s *SlaveOf) rdbFinished(count int64) {
	for i := 0; i < s.queueCount; i++ {
		go s.queueProcess(i)
	}
}

func (s *SlaveOf) queueProcess(i int) {
	lst := s.queueLists[i]
	for {
		if s.shouldStopRunloop {
			return
		}
		if lst.Len() == 0 {
			time.Sleep(time.Millisecond * time.Duration(100))
		}
		// elem, e1 := lst.LPop()
		elem := lst.LPop()
		// if e1 != nil {
		// 	fmt.Println("lpop err", i, e1)
		// 	time.Sleep(time.Millisecond * time.Duration(100))
		// 	continue
		// }
		if elem == nil {
			continue
		}
		s.syncCounters.Get("buffer").Incr(-1)
		s.syncCounters.Get("proc").Incr(1)
		// cmd, e2 := s.decodeCommand(elem.([]byte))
		// if e2 != nil {
		// 	fmt.Println("decode err", i, e2)
		// 	time.Sleep(time.Millisecond * time.Duration(100))
		// 	continue
		// }
		cmd := elem.(*Command)
		conn := s.pool.Get()
		argCount := len(cmd.Args) - 1
		objs := make([]interface{}, 0, argCount)
		for i := 0; i < argCount; i++ {
			objs = append(objs, cmd.Args[i+1])
		}
		conn.Do(cmd.Name(), objs...)
		conn.Close()
	}
}

func (s *SlaveOf) encodeCommand(cmd *Command) (out []byte, err error) {
	enc := codec.NewEncoderBytes(&out, &s.msgpack_handle)
	err = enc.Encode(cmd.Args)
	return
}

func (s *SlaveOf) decodeCommand(in []byte) (cmd *Command, err error) {
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

package goredis_server

/*
slaveClient.Start()
slaveClient.Stop()
*/
import (
	. "../goredis"
	"./libs/levelredis"
	qp "./libs/queueprocess"
	"errors"
	"fmt"
	"github.com/latermoon/levigo"
	"github.com/latermoon/msgpackgo/codec"
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
	msgpack_handle    codec.MsgpackHandle
}

func NewSlaveSessionClient(server *GoRedisServer, session *SlaveSession) (s *SlaveSessionClient) {
	s = &SlaveSessionClient{}
	s.server = server
	s.session = session
	// s.taskqueue = qp.NewQueueProcess(100, s.queueHandler)
	s.queueCount = 100
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
	db, e1 := levigo.Open(s.server.directory+"/slaveof_"+s.session.MasterHost(), opts)
	if e1 != nil {
		return e1
	}
	s.aofRedis = levelredis.NewLevelRedis(db)
	return
}

func (s *SlaveSessionClient) Start() {
	if s.shouldStopRunloop {
		s.server.stdlog.Error("[%s] slaveof should run once", s.session.RemoteAddr())
		return
	}
	for i := 0; i < s.queueCount; i++ {
		aofkey := fmt.Sprintf("queue_%d", i)
		go s.queueProcess(aofkey)
	}
	// 阻塞处理，直到出错
	s.session.DidRecvCommand = s.didRecvCommand
	err := s.session.Sync(s.server.UID())
	if err != nil {
		s.server.stdlog.Error("[%s] slaveof sync error %s", s.session.RemoteAddr(), err)
	}
	// 终止运行
	s.shouldStopRunloop = true
}

func (s *SlaveSessionClient) didRecvCommand(cmd *Command, count int64) {
	if len(cmd.Args) == 1 {
		return
	}
	key, _ := cmd.ArgAtIndex(1)
	aofkey := fmt.Sprintf("queue_%d", SumOfBytesChars(key)%s.queueCount)
	lst := s.aofRedis.GetList(aofkey)
	out, err := s.encodeCommand(cmd)
	if err == nil {
		lst.RPush(out)
	} else {
		fmt.Println("err,", err)
	}
	// _ = s.server.InvokeCommandHandler(s.session, cmd)
}

func (s *SlaveSessionClient) queueProcess(aofkey string) {
	lst := s.aofRedis.GetList(aofkey)
	for {
		if s.shouldStopRunloop {
			return
		}
		if lst.Len() == 0 {
			time.Sleep(time.Millisecond * time.Duration(100))
		}
		elem, e1 := lst.LPop()
		if e1 != nil {
			fmt.Println("lpop err", aofkey, e1)
			time.Sleep(time.Millisecond * time.Duration(100))
			continue
		}
		if elem == nil {
			continue
		}
		cmd, e2 := s.decodeCommand(elem.Value.([]byte))
		if e2 != nil {
			fmt.Println("decode err", aofkey, e1)
			time.Sleep(time.Millisecond * time.Duration(100))
			continue
		}
		fmt.Println(aofkey, lst.Len(), "->pop", cmd)
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

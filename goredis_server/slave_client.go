package goredis_server

import (
	. "GoRedis/goredis"
	"GoRedis/libs/counter"
	"GoRedis/libs/iotool"
	"GoRedis/libs/rdb"
	"GoRedis/libs/statlog"
	"GoRedis/libs/stdlog"
	"bufio"
	"errors"
	"fmt"
	"io"
	"net"
	"os"
	"strconv"
	"strings"
	"sync"
)

var slavelog = stdlog.Log("slaveof")

type SlaveClient struct {
	session  *Session
	server   *GoRedisServer
	buffer   chan *Command // 缓存实时指令
	rdbjobs  chan int      // 并发工作
	wg       sync.WaitGroup
	broken   bool // 无效连接
	counters *counter.Counters
	synclog  *statlog.StatLogger
}

func NewSlaveClient(server *GoRedisServer, session *Session) (s *SlaveClient, err error) {
	s = &SlaveClient{}
	s.server = server
	s.session = session
	s.buffer = make(chan *Command, 1000*10000)
	s.rdbjobs = make(chan int, 10)
	s.counters = counter.NewCounters()
	os.Mkdir(s.directory(), os.ModePerm)
	err = s.initLog()
	return
}

func (s *SlaveClient) initLog() error {
	path := fmt.Sprintf("%s/sync.log", s.directory())
	file, err := os.OpenFile(path, os.O_RDWR|os.O_APPEND|os.O_CREATE, os.ModePerm)
	if err != nil {
		return err
	}
	s.synclog = statlog.NewStatLogger(file)
	s.synclog.Add(statlog.TimeItem("time"))
	s.synclog.Add(statlog.Item("rdb", func() interface{} {
		return s.counters.Get("rdb").ChangedCount()
	}, &statlog.Opt{Padding: 8}))
	s.synclog.Add(statlog.Item("recv", func() interface{} {
		return s.counters.Get("recv").ChangedCount()
	}, &statlog.Opt{Padding: 8}))
	s.synclog.Add(statlog.Item("proc", func() interface{} {
		return s.counters.Get("proc").ChangedCount()
	}, &statlog.Opt{Padding: 8}))
	s.synclog.Add(statlog.Item("buffer", func() interface{} {
		return len(s.buffer)
	}, &statlog.Opt{Padding: 10}))
	go s.synclog.Start()
	return nil
}

func (s *SlaveClient) RemoteAddr() net.Addr {
	return s.session.RemoteAddr()
}

func (s *SlaveClient) directory() string {
	return s.server.directory + "sync_" + fmt.Sprint(s.session.RemoteAddr()) + "/"
}

func (s *SlaveClient) rdbfilename() string {
	return s.directory() + "dump.rdb"
}

// 开始同步
func (s *SlaveClient) Sync(uid string) (err error) {
	isgoredis, version, e1 := s.masterInfo()
	if e1 != nil {
		return e1
	}
	if isgoredis {
		slavelog.Printf("[M %s] slaveof %s GoRedis:%s\n", s.RemoteAddr(), s.RemoteAddr(), version)
	} else {
		slavelog.Printf("[M %s] slaveof %s Redis:%s\n", s.RemoteAddr(), s.RemoteAddr(), version)
	}

	args := [][]byte{[]byte("SYNC")}
	if isgoredis && len(uid) > 0 {
		args = append(args, []byte(uid))
	}
	s.session.WriteCommand(NewCommand(args...))

	rdbsaved := false
	for {
		var c byte
		c, err = s.session.PeekByte()
		// GoRedis不会发送rdb，所以直接调用recvCmd
		if isgoredis {
			go s.recvCmd()
		}
		if !rdbsaved && c == '$' {
			err = s.recvRdb()
			if err != nil {
				slavelog.Printf("[M %s] recv rdb error:%s\n", s.RemoteAddr(), err)
				break
			}
			rdbsaved = true
		} else if c == '\n' {
			s.session.ReadByte()
			s.IdleCallback()
		} else {
			var cmd *Command
			cmd, err = s.session.ReadCommand()
			if err != nil {
				break
			}
			s.CommandRecvCallback(cmd)
		}
	}
	// 跳出循环必定有错误
	s.Destory()
	return
}

func (s *SlaveClient) recvCmd() {
	for {
		if s.broken {
			break
		}
		cmd, ok := <-s.buffer
		if !ok {
			s.Destory()
			break
		}
		s.counters.Get("proc").Incr(1)
		// slavelog.Printf("[M %s] cmd: %s\n", s.RemoteAddr(), cmd)
		s.server.On(s.session, cmd)
	}
}

func (s *SlaveClient) recvRdb() (err error) {
	var f *os.File
	f, err = os.OpenFile(s.rdbfilename(), os.O_CREATE|os.O_RDWR|os.O_TRUNC, os.ModePerm)
	if err != nil {
		return
	}
	slavelog.Printf("[M %s] create rdb:%s\n", s.RemoteAddr(), s.rdbfilename())

	s.session.ReadByte()
	var size int64
	size, err = s.session.ReadInt64()
	if err != nil {
		return
	}
	s.RdbSizeCallback(size)

	// read
	w := bufio.NewWriter(f)
	// var written int64
	_, err = iotool.RateLimitCopy(w, io.LimitReader(s.session, size), 40*1024*1024, func(written int64, rate int) {
		s.RdbRecvProcessCallback(written, rate)
	})
	// _, err = io.CopyN(w, s.session, size)
	if err != nil {
		return
	}
	w.Flush()
	f.Seek(0, 0)
	// 不阻塞进行接收command
	go func() {
		s.RdbRecvFinishCallback(bufio.NewReader(f))
		filename := f.Name()
		f.Close()
		os.Remove(filename)
	}()
	return
}

// 清空本地的同步状态
func (s *SlaveClient) Destory() (err error) {
	s.broken = true
	s.synclog.Stop()
	return
}

func (s *SlaveClient) rdbFileWriter() (w *bufio.Writer, err error) {
	var file *os.File
	file, err = os.OpenFile(fmt.Sprintf("/tmp/%s.rdb", s.session.RemoteAddr()), os.O_CREATE|os.O_TRUNC|os.O_WRONLY, os.ModePerm)
	w = bufio.NewWriter(file)
	return
}

func (s *SlaveClient) masterInfo() (isgoredis bool, version string, err error) {
	cmdinfo := NewCommand([]byte("info"), []byte("server"))
	s.session.WriteCommand(cmdinfo)
	var reply *Reply
	reply, err = s.session.ReadReply()
	if err != nil {
		return
	}
	if reply.Value == nil {
		err = errors.New("reply nil")
		return
	}

	var info string
	switch reply.Value.(type) {
	case string:
		info = reply.Value.(string)
	case []byte:
		info = string(reply.Value.([]byte))
	default:
		info = reply.String()
	}

	// 切分info返回的数据，存放到map里
	kv := make(map[string]string)
	lines := strings.Split(info, "\n")
	for _, line := range lines {
		line = strings.TrimSuffix(line, "\r")
		line = strings.TrimPrefix(line, " ")
		if strings.HasPrefix(line, "#") {
			continue
		}
		pairs := strings.Split(line, ":")
		if len(pairs) != 2 {
			continue
		}
		// done
		kv[pairs[0]] = pairs[1]
	}

	_, isgoredis = kv["goredis_version"]
	if isgoredis {
		version = kv["goredis_version"]
	} else {
		version = kv["redis_version"]
	}

	return
}

func (s *SlaveClient) RdbSizeCallback(totalsize int64) {
	slavelog.Printf("[M %s] rdb size: %d\n", s.RemoteAddr(), totalsize)
}

func (s *SlaveClient) RdbRecvFinishCallback(r *bufio.Reader) {
	slavelog.Printf("[M %s] rdb recv finish, start decoding... \n", s.RemoteAddr())
	// decode
	dec := newRdbDecoder(s)
	err := rdb.Decode(r, dec)
	if err != nil {
		// must cancel
		slavelog.Printf("[M %s] decode error %s\n", s.RemoteAddr(), err)
		s.Destory()
	}
	return
}

func (s *SlaveClient) rdbDecodeCommand(cmd *Command) {
	// slavelog.Printf("[M %s] rdb decode %s\n", client.RemoteAddr(), cmd)
	s.counters.Get("rdb").Incr(1)
	s.rdbjobs <- 1
	s.wg.Add(1)
	go func() {
		s.server.On(s.session, cmd)
		<-s.rdbjobs
		s.wg.Done()
	}()
}

func (s *SlaveClient) rdbDecodeFinish(n int64) {
	slavelog.Printf("[M %s] rdb decode finish, items: %d\n", s.RemoteAddr(), n)
	s.wg.Wait()
	go s.recvCmd() // 开始消化command
}

func (s *SlaveClient) RdbRecvProcessCallback(size int64, rate int) {
	slavelog.Printf("[M %s] rdb recv: %d, rate:%d\n", s.RemoteAddr(), size, rate)
}

func (s *SlaveClient) IdleCallback() {
	slavelog.Printf("[M %s] slaveof waiting\n", s.RemoteAddr())
}

func (s *SlaveClient) CommandRecvCallback(cmd *Command) {
	// slavelog.Printf("[M %s] recv: %s\n", s.RemoteAddr(), cmd)
	s.counters.Get("recv").Incr(1)
	s.buffer <- cmd
}

// =============================================
// 第三方rdb解释函数
// =============================================
type rdbDecoder struct {
	rdb.NopDecoder
	db       int
	i        int
	keyCount int64
	bufsize  int
	client   *SlaveClient
	// 数据缓冲
	hashEntry [][]byte
	setEntry  [][]byte
	listEntry [][]byte
	zsetEntry [][]byte
}

func newRdbDecoder(s *SlaveClient) (dec *rdbDecoder) {
	dec = &rdbDecoder{}
	dec.client = s
	dec.keyCount = 0
	dec.bufsize = 200
	return
}

func (p *rdbDecoder) StartDatabase(n int) {
	p.db = n
}

func (p *rdbDecoder) EndDatabase(n int) {
}

func (p *rdbDecoder) EndRDB() {
	p.client.rdbDecodeFinish(p.keyCount)
}

// Set
func (p *rdbDecoder) Set(key, value []byte, expiry int64) {
	cmd := NewCommand([]byte("SET"), key, value)
	p.client.rdbDecodeCommand(cmd)
	p.keyCount++
}

func (p *rdbDecoder) StartHash(key []byte, length, expiry int64) {
	if int(length) < p.bufsize {
		p.hashEntry = make([][]byte, 0, length+2)
	} else {
		p.hashEntry = make([][]byte, 0, p.bufsize)
	}
	p.hashEntry = append(p.hashEntry, []byte("HMSET"))
	p.hashEntry = append(p.hashEntry, key)
	p.keyCount++
}

func (p *rdbDecoder) Hset(key, field, value []byte) {
	p.hashEntry = append(p.hashEntry, field)
	p.hashEntry = append(p.hashEntry, value)
	if len(p.hashEntry) >= p.bufsize {
		cmd := NewCommand(p.hashEntry...)
		p.client.rdbDecodeCommand(cmd)
		p.hashEntry = make([][]byte, 0, p.bufsize)
		p.hashEntry = append(p.hashEntry, []byte("HMSET"))
		p.hashEntry = append(p.hashEntry, key)
	}
}

// Hash
func (p *rdbDecoder) EndHash(key []byte) {
	if len(p.hashEntry) > 2 {
		cmd := NewCommand(p.hashEntry...)
		p.client.rdbDecodeCommand(cmd)
	}
}

func (p *rdbDecoder) StartSet(key []byte, cardinality, expiry int64) {
	if int(cardinality) < p.bufsize {
		p.setEntry = make([][]byte, 0, cardinality+2)
	} else {
		p.setEntry = make([][]byte, 0, p.bufsize)
	}
	p.setEntry = append(p.setEntry, []byte("SADD"))
	p.setEntry = append(p.setEntry, key)
	p.keyCount++
}

func (p *rdbDecoder) Sadd(key, member []byte) {
	p.setEntry = append(p.setEntry, member)
	if len(p.setEntry) >= p.bufsize {
		cmd := NewCommand(p.setEntry...)
		p.client.rdbDecodeCommand(cmd)
		p.setEntry = make([][]byte, 0, p.bufsize)
		p.setEntry = append(p.setEntry, []byte("SADD"))
		p.setEntry = append(p.setEntry, key)
	}
}

// Set
func (p *rdbDecoder) EndSet(key []byte) {
	if len(p.setEntry) > 2 {
		cmd := NewCommand(p.setEntry...)
		p.client.rdbDecodeCommand(cmd)
	}
}

func (p *rdbDecoder) StartList(key []byte, length, expiry int64) {
	if int(length) < p.bufsize {
		p.listEntry = make([][]byte, 0, length+2)
	} else {
		p.listEntry = make([][]byte, 0, p.bufsize)
	}
	p.listEntry = append(p.listEntry, []byte("RPUSH"))
	p.listEntry = append(p.listEntry, key)
	p.keyCount++
	p.i = 0
}

func (p *rdbDecoder) Rpush(key, value []byte) {
	p.listEntry = append(p.listEntry, value)
	if len(p.listEntry) >= p.bufsize {
		cmd := NewCommand(p.listEntry...)
		p.client.rdbDecodeCommand(cmd)
		p.listEntry = make([][]byte, 0, p.bufsize)
		p.listEntry = append(p.listEntry, []byte("RPUSH"))
		p.listEntry = append(p.listEntry, key)
	}
	p.i++
}

// List
func (p *rdbDecoder) EndList(key []byte) {
	if len(p.listEntry) > 2 {
		cmd := NewCommand(p.listEntry...)
		p.client.rdbDecodeCommand(cmd)
	}
}

func (p *rdbDecoder) StartZSet(key []byte, cardinality, expiry int64) {
	if int(cardinality) > p.bufsize {
		p.zsetEntry = make([][]byte, 0, cardinality)
	} else {
		p.zsetEntry = make([][]byte, 0, p.bufsize)
	}
	p.zsetEntry = append(p.zsetEntry, []byte("ZADD"))
	p.zsetEntry = append(p.zsetEntry, key)
	p.keyCount++
	p.i = 0
}

func (p *rdbDecoder) Zadd(key []byte, score float64, member []byte) {
	p.zsetEntry = append(p.zsetEntry, []byte(strconv.FormatInt(int64(score), 10)))
	p.zsetEntry = append(p.zsetEntry, member)
	if len(p.zsetEntry) >= p.bufsize {
		cmd := NewCommand(p.zsetEntry...)
		p.client.rdbDecodeCommand(cmd)
		p.zsetEntry = make([][]byte, 0, p.bufsize)
		p.zsetEntry = append(p.zsetEntry, []byte("ZADD"))
		p.zsetEntry = append(p.zsetEntry, key)
	}
	p.i++
}

// ZSet
func (p *rdbDecoder) EndZSet(key []byte) {
	if len(p.zsetEntry) > 2 {
		cmd := NewCommand(p.zsetEntry...)
		p.client.rdbDecodeCommand(cmd)
	}
}

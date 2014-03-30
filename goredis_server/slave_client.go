package goredis_server

// 管理同步连接，从master获取数据更新到本地
import (
	. "GoRedis/goredis"
	"GoRedis/libs/counter"
	"GoRedis/libs/iotool"
	"GoRedis/libs/rdb"
	"GoRedis/libs/stat"
	"GoRedis/libs/stdlog"
	"bufio"
	"fmt"
	"io"
	"os"
	"strconv"
	"sync"
)

var slavelog = stdlog.Log("slaveof")

type SlaveClient struct {
	ISlaveClient
	session  *Session
	server   *GoRedisServer
	buffer   chan *Command // 缓存实时指令
	rdbjobs  chan int      // 并发工作
	wg       sync.WaitGroup
	broken   bool // 无效连接
	counters *counter.Counters
	synclog  *stat.Writer
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
	s.synclog = stat.New(file)
	st := s.synclog
	st.Add(stat.TextItem("time", 8, func() interface{} { return stat.TimeString() }))
	st.Add(stat.IncrItem("rdb", 8, func() int64 { return s.counters.Get("rdb").Count() }))
	st.Add(stat.IncrItem("recv", 8, func() int64 { return s.counters.Get("recv").Count() }))
	st.Add(stat.IncrItem("proc", 8, func() int64 { return s.counters.Get("proc").Count() }))
	st.Add(stat.TextItem("buffer", 10, func() interface{} { return len(s.buffer) }))
	go st.Start()
	return nil
}

func (s *SlaveClient) Session() *Session {
	return s.session
}

func (s *SlaveClient) directory() string {
	return s.server.directory + "sync_" + fmt.Sprint(s.session.RemoteAddr()) + "/"
}

func (s *SlaveClient) rdbfilename() string {
	return s.directory() + "dump.rdb"
}

// 开始同步
func (s *SlaveClient) Sync() (err error) {

	err = s.session.WriteCommand(NewCommand([]byte("SYNC")))
	if err != nil {
		return
	}

	rdbsaved := false
	for {
		var c byte
		c, err = s.session.PeekByte()
		if err != nil {
			break
		}
		if !rdbsaved && c == '$' {
			s.session.SetAttribute(S_STATUS, REPL_RECV_BULK)
			err = s.recvRdb()
			if err != nil {
				slavelog.Printf("[M %s] recv rdb error:%s\n", s.session.RemoteAddr(), err)
				break
			}
			rdbsaved = true
		} else if c == '\n' {
			_, err = s.session.ReadByte()
			if err != nil {
				break
			}
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
	s.Close()
	return
}

func (s *SlaveClient) Close() {
	s.broken = true
	s.synclog.Close()
	s.session.Close()
	return
}

func (s *SlaveClient) recvCmd() {
	for {
		if s.broken {
			break
		}
		cmd, ok := <-s.buffer
		if !ok {
			break
		}
		s.counters.Get("proc").Incr(1)
		s.server.On(s.session, cmd)
	}
}

func (s *SlaveClient) recvRdb() (err error) {
	var f *os.File
	f, err = os.OpenFile(s.rdbfilename(), os.O_CREATE|os.O_RDWR|os.O_TRUNC, os.ModePerm)
	if err != nil {
		return
	}
	slavelog.Printf("[M %s] create rdb:%s\n", s.session.RemoteAddr(), s.rdbfilename())

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

func (s *SlaveClient) rdbFileWriter() (w *bufio.Writer, err error) {
	var file *os.File
	file, err = os.OpenFile(fmt.Sprintf("/tmp/%s.rdb", s.session.RemoteAddr()), os.O_CREATE|os.O_TRUNC|os.O_WRONLY, os.ModePerm)
	w = bufio.NewWriter(file)
	return
}

func (s *SlaveClient) RdbSizeCallback(totalsize int64) {
	slavelog.Printf("[M %s] rdb size: %d\n", s.session.RemoteAddr(), totalsize)
}

func (s *SlaveClient) RdbRecvFinishCallback(r *bufio.Reader) {
	slavelog.Printf("[M %s] rdb recv finish, start decoding... \n", s.session.RemoteAddr())
	// decode
	dec := newRdbDecoder(s)
	err := rdb.Decode(r, dec)
	if err != nil {
		// must cancel
		slavelog.Printf("[M %s] decode error %s\n", s.session.RemoteAddr(), err)
		s.Close()
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
	slavelog.Printf("[M %s] rdb decode finish, items: %d\n", s.session.RemoteAddr(), n)
	s.session.SetAttribute(S_STATUS, REPL_ONLINE)
	s.wg.Wait()
	go s.recvCmd() // 开始消化command
}

func (s *SlaveClient) RdbRecvProcessCallback(size int64, rate int) {
	slavelog.Printf("[M %s] rdb recv: %d, rate:%d\n", s.session.RemoteAddr(), size, rate)
}

func (s *SlaveClient) IdleCallback() {
	slavelog.Printf("[M %s] slaveof waiting\n", s.session.RemoteAddr())
}

func (s *SlaveClient) CommandRecvCallback(cmd *Command) {
	// slavelog.Printf("[M %s] recv: %s\n", s.session.RemoteAddr(), cmd)
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

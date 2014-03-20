package slaveof

import (
	. "GoRedis/goredis"
	"GoRedis/libs/counter"
	"GoRedis/libs/iotool"
	"GoRedis/libs/ratelimit"
	"GoRedis/libs/rdb"
	"GoRedis/libs/stat"
	"GoRedis/libs/stdlog"
	"bufio"
	"fmt"
	"io"
	"net"
	"os"
	"sync"
	"time"
)

type SlaveClient struct {
	srchost    string
	desthost   string
	src        *Session          // 主库
	dest       *Session          // 从库
	destwriter io.Writer         //
	directory  string            // 工作目录
	buffer     chan *Command     // 缓存实时指令
	jobs       chan int          // 并发工作
	wg         sync.WaitGroup    //
	counters   *counter.Counters //
	synclog    *stat.Writer      //
	totalsize  int64             //
	pullrate   int               // rdb传输速率
	pushrate   int               //
	mu         sync.Mutex        //
	connmu     sync.Mutex        //
	online     bool              //
}

func NewClient(srchost, desthost string, buffersize int) (s *SlaveClient, err error) {
	s = &SlaveClient{
		srchost:   srchost,
		desthost:  desthost,
		directory: "/tmp",
		buffer:    make(chan *Command, buffersize*10000),
		jobs:      make(chan int, 10),
		counters:  counter.NewCounters(),
		pullrate:  40 * 1024 * 1024, // 40MB
		pushrate:  40 * 1024 * 1024, // 40MB
	}
	// warmup
	conn, e1 := net.Dial("tcp", s.srchost)
	if e1 != nil {
		return nil, e1
	}
	s.src = NewSession(conn)
	return
}

func (s *SlaveClient) initlog() error {
	s.synclog = stat.New(os.Stdout)
	st := s.synclog
	st.Add(stat.TextItem("time", 8, func() interface{} { return stat.TimeString() }))
	st.Add(stat.IncrItem("rdb", 8, func() int64 { return s.counters.Get("rdb").Count() }))
	st.Add(stat.IncrItem("in", 8, func() int64 { return s.counters.Get("in").Count() }))
	st.Add(stat.IncrItem("out", 8, func() int64 { return s.counters.Get("out").Count() }))
	st.Add(stat.TextItem("buffer", 10, func() interface{} { return len(s.buffer) }))
	go st.Start()
	return nil
}

func (s *SlaveClient) SetPullRate(n int) {
	s.pullrate = n
}

func (s *SlaveClient) SetPushRate(n int) {
	s.pushrate = n
}

func (s *SlaveClient) Writer() io.Writer {
	if s.destwriter == nil {
		s.destwriter = ratelimit.NewRateLimiter(s.Dest(), s.pushrate)
	}
	return s.destwriter
}

func (s *SlaveClient) Dest() *Session {
	s.connmu.Lock()
	defer s.connmu.Unlock()
	if s.dest == nil {
		for {
			conn, err := net.Dial("tcp", s.desthost)
			if err != nil {
				stdlog.Println("CONN_ERR", err)
				time.Sleep(time.Millisecond * 1000)
				continue
			}
			s.dest = NewSession(conn)
			break
		}
	}
	return s.dest
}

func (s *SlaveClient) rdbfilename() string {
	return fmt.Sprintf("%s/%s_dump.db", s.directory, s.src.RemoteAddr())
}

func (s *SlaveClient) Sync() (err error) {
	session := s.src
	if err = session.WriteCommand(NewCommand([]byte("SYNC"))); err != nil {
		return
	}

	rdbsaved := false
	for {
		var c byte
		if c, err = session.PeekByte(); err != nil {
			break
		}
		if !rdbsaved && c == '$' {
			if err = s.recvRdb(); err != nil {
				break
			}
			rdbsaved = true
		} else if c == '\n' {
			if _, err = session.ReadByte(); err != nil {
				break
			}
			stdlog.Println("waiting ...")
		} else {
			var cmd *Command
			if cmd, err = session.ReadCommand(); err != nil {
				break
			}
			s.counters.Get("in").Incr(1)
			s.buffer <- cmd
		}
	}
	return
}

func (s *SlaveClient) Close() {
	s.synclog.Close()
	s.src.Close()
	s.Dest().Close()
}

func (s *SlaveClient) procCommand() {
	for {
		cmd, ok := <-s.buffer
		if !ok {
			break
		}
		s.counters.Get("out").Incr(1)
		s.writeCommand(cmd)
	}
}

func (s *SlaveClient) writeCommand(cmd *Command) {
	s.mu.Lock()
	defer s.mu.Unlock()
	for {
		_, err := s.Writer().Write(cmd.Bytes())
		if err != nil {
			s.dest = nil
			stdlog.Println("WRITE_ERR", err, cmd)
			time.Sleep(time.Millisecond * 1000)
			continue
		}
		break
	}
}

func (s *SlaveClient) readAllReply() {
	for {
		_, err := s.Dest().ReadByte()
		if err != nil {
			s.dest = nil
			stdlog.Println("ReadReply ERR", err)
			time.Sleep(time.Millisecond * 1000)
		}
	}
}

func (s *SlaveClient) recvRdb() (err error) {
	session := s.src
	var f *os.File
	f, err = os.OpenFile(s.rdbfilename(), os.O_CREATE|os.O_RDWR|os.O_TRUNC, os.ModePerm)
	if err != nil {
		return
	}
	stdlog.Printf("[M %s] create rdb:%s\n", session.RemoteAddr(), s.rdbfilename())

	session.ReadByte()
	var size int64
	size, err = session.ReadInt64()
	if err != nil {
		return
	}
	s.RdbSizeCallback(size)

	// read
	w := bufio.NewWriter(f)
	// var written int64
	_, err = iotool.RateLimitCopy(w, io.LimitReader(session, size), s.pullrate, func(written int64, rate int) {
		s.RdbRecvProcessCallback(written, rate)
	})
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

func (s *SlaveClient) RdbSizeCallback(totalsize int64) {
	s.totalsize = totalsize
	stdlog.Printf("[M %s] rdb size: %s\n", s.src.RemoteAddr(), bytesInHuman(totalsize))
}

func (s *SlaveClient) RdbRecvFinishCallback(r *bufio.Reader) {
	stdlog.Printf("[M %s] rdb recv finish, start decoding... \n", s.src.RemoteAddr())
	s.initlog()
	go s.readAllReply()
	// decode
	dec := newRdbDecoder(s)
	err := rdb.Decode(r, dec)
	if err != nil {
		// must cancel
		stdlog.Printf("[M %s] decode error %s\n", s.src.RemoteAddr(), err)
		s.Close()
	}
	return
}

func (s *SlaveClient) rdbDecodeCommand(cmd *Command) {
	s.counters.Get("rdb").Incr(1)
	s.jobs <- 1
	s.wg.Add(1)
	go func() {
		s.writeCommand(cmd)
		<-s.jobs
		s.wg.Done()
	}()
}

func (s *SlaveClient) rdbDecodeFinish(n int64) {
	// stdlog.Printf("[M %s] rdb decode finish, items: %d\n", s.src.RemoteAddr(), n)
	s.wg.Wait()
	s.online = true
	go s.procCommand() // 开始消化command
}

func (s *SlaveClient) RdbRecvProcessCallback(size int64, rate int) {
	stdlog.Printf("[M %s] rdb recv: %s/%s, rate:%s\n", s.src.RemoteAddr(), bytesInHuman(size), bytesInHuman(s.totalsize), bytesInHuman(int64(rate)))
}

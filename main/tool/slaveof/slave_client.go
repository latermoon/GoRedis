package slaveof

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
	"net"
	"os"
	"sync"
)

type SlaveClient struct {
	src       *Session          // 主库
	dest      *Session          // 从库
	directory string            // 工作目录
	buffer    chan *Command     // 缓存实时指令
	jobs      chan int          // 并发工作
	wg        sync.WaitGroup    //
	counters  *counter.Counters //
	synclog   *stat.Writer      //
	rdbrate   int               // rdb传输速率
}

func NewClient(src net.Conn, dest net.Conn) (s *SlaveClient) {
	s = &SlaveClient{
		src:       NewSession(src),
		dest:      NewSession(dest),
		directory: "/tmp/",
		buffer:    make(chan *Command, 1000*10000),
		jobs:      make(chan int, 10),
		counters:  counter.NewCounters(),
		rdbrate:   40 * 1024 * 1024, // 40MB
	}
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
	s.rdbrate = n
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
	s.dest.Close()
}

func (s *SlaveClient) procCommand() {
	for {
		cmd, ok := <-s.buffer
		if !ok {
			break
		}
		s.counters.Get("out").Incr(1)
		s.dest.WriteCommand(cmd)
	}
}

func (s *SlaveClient) cleanReply() {
	for {
		_, err := s.dest.ReadByte()
		if err != nil {
			stdlog.Println("cleanReply error", err)
			break
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
	_, err = iotool.RateLimitCopy(w, io.LimitReader(session, size), s.rdbrate, func(written int64, rate int) {
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
	stdlog.Printf("[M %s] rdb size: %d\n", s.src.RemoteAddr(), totalsize)
}

func (s *SlaveClient) RdbRecvFinishCallback(r *bufio.Reader) {
	stdlog.Printf("[M %s] rdb recv finish, start decoding... \n", s.src.RemoteAddr())
	s.initlog()
	go s.cleanReply()
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
		s.dest.WriteCommand(cmd)
		<-s.jobs
		s.wg.Done()
	}()
}

func (s *SlaveClient) rdbDecodeFinish(n int64) {
	// stdlog.Printf("[M %s] rdb decode finish, items: %d\n", s.src.RemoteAddr(), n)
	s.wg.Wait()
	go s.procCommand() // 开始消化command
}

func (s *SlaveClient) RdbRecvProcessCallback(size int64, rate int) {
	stdlog.Printf("[M %s] rdb recv: %d, rate:%d\n", s.src.RemoteAddr(), size, rate)
}

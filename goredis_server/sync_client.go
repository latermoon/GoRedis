package goredis_server

import (
	. "GoRedis/goredis"
	"GoRedis/libs/counter"
	"GoRedis/libs/statlog"
	"GoRedis/libs/stdlog"
	"errors"
	"fmt"
	"os"
	"sync"
)

var (
	SyncError            = errors.New("sync errors")
	SyncOutOfBufferError = errors.New("sync out of buffer")
)

// 负责传输数据到从库
// status = none/connected/disconnect
type SyncClient struct {
	session     *Session
	home        string
	buffer      chan *Command
	synclog     *statlog.StatLogger
	counters    *counter.Counters
	unavailable bool // 连接不可用
	mu          sync.Mutex
	status      string
}

func NewSyncClient(session *Session, home string) (s *SyncClient, err error) {
	s = &SyncClient{}
	s.session = session
	s.home = home
	s.buffer = make(chan *Command, 500*10000)
	s.counters = counter.NewCounters()
	// log
	logfile := fmt.Sprintf("%s/slave_%s.log", home, session.RemoteAddr())
	if err = s.initSyncLog(logfile); err != nil {
		return
	}
	s.status = "sendbulk"
	return
}

func (s *SyncClient) initSyncLog(filename string) error {
	if file, err := os.OpenFile(filename, os.O_CREATE|os.O_APPEND|os.O_WRONLY, os.ModePerm); err != nil {
		return err
	} else {
		s.synclog = statlog.NewStatLogger(file)
	}
	s.synclog.Add(statlog.TimeItem("time"))
	s.synclog.Add(statlog.Item("in", func() interface{} {
		return s.counters.Get("in").ChangedCount()
	}, &statlog.Opt{Padding: 10}))
	s.synclog.Add(statlog.Item("out", func() interface{} {
		return s.counters.Get("out").ChangedCount()
	}, &statlog.Opt{Padding: 10}))
	s.synclog.Add(statlog.Item("buffer", func() interface{} {
		return len(s.buffer)
	}, &statlog.Opt{Padding: 10}))
	return nil
}

func (s *SyncClient) Status() string {
	return s.status
}

func (s *SyncClient) Available() bool {
	return !s.unavailable
}

// 将要发送的指令放入待传输队列
func (s *SyncClient) Enqueue(cmd *Command) (err error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.unavailable {
		return SyncError
	}
	if len(s.buffer) == cap(s.buffer) {
		return SyncOutOfBufferError
	}
	s.counters.Get("in").Incr(1)
	s.buffer <- cmd
	return
}

func (s *SyncClient) Send(cmd *Command) (err error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.unavailable {
		return SyncError
	}
	s.counters.Get("out").Incr(1)
	err = s.session.WriteCommand(cmd)
	if err != nil {
		stdlog.Printf("sync error %s %s\n", s.session.RemoteAddr(), err)
		s.cancel()
	}
	return
}

// 开始同步
func (s *SyncClient) StartSync() {
	s.status = "online"
	go func() {
		for {
			if s.unavailable {
				break
			}
			cmd, ok := <-s.buffer
			if !ok {
				break
			}
			if err := s.Send(cmd); err != nil {
				break
			}
		}
		stdlog.Printf("[%s] syncloop closed", s.session.RemoteAddr())
	}()
}

func (s *SyncClient) Cancel() {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.cancel()
}

func (s *SyncClient) cancel() {
	s.status = "offline"
	stdlog.Printf("[%s] sync cancel\n", s.session.RemoteAddr())
	s.unavailable = true
	s.synclog.Stop()
	close(s.buffer)
}

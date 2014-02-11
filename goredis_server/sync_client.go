package goredis_server

import (
	"./monitor"
	. "GoRedis/goredis"
	"GoRedis/libs/statlog"
	"GoRedis/libs/stdlog"
	"fmt"
	"os"
)

type SyncClient struct {
	session     *Session
	home        string
	buffer      chan *Command
	synclog     *statlog.StatLogger
	counters    *monitor.Counters
	unavailable bool // 连接不可用
}

func NewSyncClient(session *Session, home string) (s *SyncClient, err error) {
	s = &SyncClient{}
	s.session = session
	s.home = home
	s.buffer = make(chan *Command, 1000*10000)
	s.counters = monitor.NewCounters()
	// log
	logfile := fmt.Sprintf("%s/slave_%s.log", home, session.RemoteAddr())
	if err = s.initSyncLog(logfile); err != nil {
		return
	}
	go s.runloop()
	return
}

func (s *SyncClient) initSyncLog(filename string) error {
	if file, err := os.OpenFile(filename, os.O_CREATE|os.O_APPEND|os.O_WRONLY, os.ModePerm); err != nil {
		return err
	} else {
		s.synclog = statlog.NewStatLogger(file)
	}
	s.synclog.Add(statlog.TimeItem("time"))
	s.synclog.Add(statlog.Item("recv", func() interface{} {
		return s.counters.Get("recv").ChangedCount()
	}, &statlog.Opt{Padding: 10}))
	s.synclog.Add(statlog.Item("send", func() interface{} {
		return s.counters.Get("send").ChangedCount()
	}, &statlog.Opt{Padding: 10}))
	s.synclog.Add(statlog.Item("buffer", func() interface{} {
		return len(s.buffer)
	}, &statlog.Opt{Padding: 10}))
	return nil
}

func (s *SyncClient) SendCommand(cmd *Command) {
	if s.unavailable {
		return
	}
	s.counters.Get("recv").Incr(1)
	s.buffer <- cmd
}

// 将buffer中的command发出去
func (s *SyncClient) runloop() {
	for {
		cmd := <-s.buffer
		s.counters.Get("send").Incr(1)
		err := s.session.WriteCommand(cmd)
		if err != nil {
			s.unavailable = true
			stdlog.Printf("sync error %s %s\n", s.session.RemoteAddr(), err)
			s.synclog.Stop()
			break
		}
	}
}

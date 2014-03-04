package goredis_server

import (
	. "GoRedis/goredis"
	"GoRedis/libs/counter"
	"GoRedis/libs/statlog"
	"fmt"
	"net"
	"os"
)

type SlaveClientV2 struct {
	ISlaveClient
	session  *Session
	server   *GoRedisServer
	broken   bool
	status   string
	lastseq  int64
	counters *counter.Counters
	synclog  *statlog.StatLogger
}

func NewSlaveClientV2(server *GoRedisServer, session *Session) (s *SlaveClientV2, err error) {
	s = &SlaveClientV2{
		server:   server,
		session:  session,
		counters: counter.NewCounters(),
	}
	err = s.initLog()
	return
}

func (s *SlaveClientV2) initLog() error {
	path := fmt.Sprintf("%s/sync_%s.log", s.server.directory, s.session.RemoteAddr())
	file, err := os.OpenFile(path, os.O_RDWR|os.O_APPEND|os.O_CREATE, os.ModePerm)
	if err != nil {
		return err
	}
	s.synclog = statlog.NewStatLogger(file)
	s.synclog.Add(statlog.TimeItem("time"))

	s.synclog.Add(statlog.Item("raw", func() interface{} {
		return s.counters.Get("raw").ChangedCount()
	}, &statlog.Opt{Padding: 8}))

	s.synclog.Add(statlog.Item("cmd", func() interface{} {
		return s.counters.Get("cmd").ChangedCount()
	}, &statlog.Opt{Padding: 8}))

	s.synclog.Add(statlog.Item("seq", func() interface{} {
		return s.lastseq
	}, &statlog.Opt{Padding: 16}))
	go s.synclog.Start()
	return nil
}

func (s *SlaveClientV2) RemoteAddr() net.Addr {
	return s.session.RemoteAddr()
}

func (s *SlaveClientV2) Broken() bool {
	return s.broken
}

func (s *SlaveClientV2) Status() string {
	return s.status
}

func (s *SlaveClientV2) Destory() {
	s.status = "broken"
	s.broken = true
	s.synclog.Stop()
}

func (s *SlaveClientV2) Sync(uid string) (err error) {
	s.status = "recv"

	s.lastseq = s.masterSeq(s.session.RemoteAddr().String())
	seq := s.lastseq + 1 // 向服务器下一个数据
	synccmd := NewCommand(formatByteSlice("SYNC", uid, seq)...)
	slavelog.Printf("[M %s] %s\n", s.session.RemoteAddr(), synccmd)

	if err = s.session.WriteCommand(synccmd); err != nil {
		slavelog.Printf("[M %s] sync error %s", s.session.RemoteAddr(), err)
		s.Destory()
		return
	}

	for {
		cmd, err := s.session.ReadCommand()
		if err != nil {
			slavelog.Printf("[M %s] master closed %s\n", s.session.RemoteAddr(), err)
			break
		}
		cmdName := cmd.Name()
		switch cmdName {
		case "SYNC_RAW_BEG":
		case "SYNC_RAW":
			s.counters.Get("raw").Incr(1)
			s.server.OnRAW_SET(cmd)
		case "SYNC_RAW_FIN":
		case "SYNC_SEQ_BEG":
			// 进入后只有出错才退出
			s.status = "online"
			s.onSYNC_SEQ_BEG(cmd)
			break
		default:
			s.server.On(s.session, cmd)
		}
	}
	s.Destory()
	return
}

// 收取快照完成后，开始收取实时数据
func (s *SlaveClientV2) onSYNC_SEQ_BEG(cmd *Command) {
	session := s.session
	for {
		// SYNC_SEQ
		cmd, err := session.ReadCommand()
		if err != nil || cmd.Name() != "SYNC_SEQ" {
			slavelog.Printf("[M %s] master closed %s %s\n", session.RemoteAddr(), cmd, err)
			break
		}

		s.lastseq, err = cmd.Int64AtIndex(1)
		if err != nil {
			slavelog.Printf("[M %s] master seq err %s\n", session.RemoteAddr(), err)
			break
		}

		cmd, err = session.ReadCommand()
		if err != nil {
			slavelog.Printf("[M %s] master closed err %s\n", session.RemoteAddr(), err)
			break
		}
		// no reply
		s.counters.Get("cmd").Incr(1)
		s.server.On(session, cmd)
		s.updateMasterSeq(session.RemoteAddr().String(), s.lastseq)
	}
	s.Destory()
	return
}

func (s *SlaveClientV2) masterSeq(host string) (seq int64) {
	key := "master:" + host + ":seq"
	seq = s.server.config.IntForKey(key, -2)
	return
}

func (s *SlaveClientV2) updateMasterSeq(host string, seq int64) {
	key := "master:" + host + ":seq"
	s.server.config.SetInt(key, seq)
}

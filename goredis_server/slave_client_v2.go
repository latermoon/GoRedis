package goredis_server

import (
	. "GoRedis/goredis"
	"GoRedis/libs/counter"
	"GoRedis/libs/stat"
	"fmt"
	"net"
	"os"
)

type ISlaveClient interface {
	Sync(uid string) (err error)
	Broken() bool
	RemoteAddr() net.Addr
	Status() string
}

type SlaveClientV2 struct {
	ISlaveClient
	session  *Session
	server   *GoRedisServer
	broken   bool
	status   string
	lastseq  int64
	counters *counter.Counters
	synclog  *stat.Writer
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
	s.synclog = stat.New(file)
	st := s.synclog
	st.Add(stat.TextItem("time", 8, func() interface{} { return stat.TimeString() }))
	st.Add(stat.IncrItem("raw", 8, func() int64 { return s.counters.Get("raw").Count() }))
	st.Add(stat.IncrItem("cmd", 8, func() int64 { return s.counters.Get("cmd").Count() }))
	st.Add(stat.TextItem("seq", 16, func() interface{} { return s.lastseq }))
	go st.Start()

	return nil
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
			slavelog.Printf("[M %s] sync raw start\n", s.session.RemoteAddr())
		case "SYNC_RAW":
			s.counters.Get("raw").Incr(1)
			s.server.OnRAW_SET(cmd)
		case "SYNC_RAW_FIN":
			slavelog.Printf("[M %s] sync raw finish\n", s.session.RemoteAddr())
		case "SYNC_SEQ_BEG":
			s.status = "online"
			slavelog.Printf("[M %s] sync cmd ...\n", s.session.RemoteAddr())
			s.recvCommandSeq(cmd) // 进入后只有出错才退出
			break
		default:
			s.server.On(s.session, cmd)
		}
	}
	s.Destory()
	return
}

// 收取快照完成后，开始收取实时数据
func (s *SlaveClientV2) recvCommandSeq(cmd *Command) (err error) {
	session := s.session
	for {
		// SYNC_SEQ
		cmd, err = session.ReadCommand()
		if err != nil {
			break
		}
		cmdName := cmd.Name()
		switch cmdName {
		case "PING":
			continue
		case "SYNC_SEQ":
			s.lastseq, err = cmd.Int64AtIndex(1)
			if err != nil {
				break
			}
		default: // commands
			s.counters.Get("cmd").Incr(1)
			s.server.On(session, cmd)
			s.updateMasterSeq(session.RemoteAddr().String(), s.lastseq)
		}
	}
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
	s.synclog.Close()
}

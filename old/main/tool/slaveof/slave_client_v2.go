// package slaveof

// import (
// 	. "GoRedis/goredis"
// 	"GoRedis/libs/counter"
// 	"GoRedis/libs/stat"
// 	"fmt"
// 	"os"
// 	"strings"
// )

// type ISlaveClient interface {
// 	Sync() (err error)
// 	Session() *Session
// 	Close()
// }

// type SlaveClientV2 struct {
// 	ISlaveClient
// 	session  *Session
// 	desthost string
// 	dest     *Session
// 	lastseq  int64
// 	counters *counter.Counters
// 	synclog  *stat.Writer
// }

// func NewSlaveClientV2(session *Session, desthost string) (s *SlaveClientV2, err error) {
// 	s = &SlaveClientV2{
// 		session:  session,
// 		desthost: desthost,
// 		counters: counter.NewCounters(),
// 	}
// 	err = s.initLog()
// 	return
// }

// func (s *SlaveClientV2) initLog() error {
// 	path := fmt.Sprintf("/tmp/sync_%s.log", s.session.RemoteAddr())
// 	file, err := os.OpenFile(path, os.O_RDWR|os.O_APPEND|os.O_CREATE, os.ModePerm)
// 	if err != nil {
// 		return err
// 	}
// 	s.synclog = stat.New(file)
// 	st := s.synclog
// 	st.Add(stat.TextItem("time", 8, func() interface{} { return stat.TimeString() }))
// 	st.Add(stat.IncrItem("raw", 8, func() int64 { return s.counters.Get("raw").Count() }))
// 	st.Add(stat.IncrItem("cmd", 8, func() int64 { return s.counters.Get("cmd").Count() }))
// 	st.Add(stat.TextItem("seq", 16, func() interface{} { return s.lastseq }))
// 	go st.Start()

// 	return nil
// }

// func (s *SlaveClientV2) Session() *Session {
// 	return s.session
// }

// func (s *SlaveClientV2) Sync() (err error) {
// 	s.lastseq = -2
// 	args := formatByteSlice("SYNC", "UID", "", "PORT", s.server.opt.Port())
// 	if s.lastseq < 0 {
// 		args = append(args, formatByteSlice("SNAP", "1")...)
// 	} else {
// 		args = append(args, formatByteSlice("SEQ", s.lastseq)...)
// 	}
// 	synccmd := NewCommand(args...)
// 	slavelog.Printf("[M %s] %s\n", s.session.RemoteAddr(), synccmd)

// 	if err = s.session.WriteCommand(synccmd); err != nil {
// 		return
// 	}

// 	for {
// 		var cmd *Command
// 		cmd, err = s.session.ReadCommand()
// 		if err != nil {
// 			break
// 		}
// 		cmdName := cmd.Name()
// 		switch cmdName {
// 		case "SYNC_RAW_START":
// 			s.Session().SetAttribute(S_STATUS, REPL_RECV_BULK)
// 			slavelog.Printf("[M %s] recv bulk start\n", s.session.RemoteAddr())
// 		case "SYNC_RAW":
// 			s.counters.Get("raw").Incr(1)
// 			s.server.OnRAW_SET(cmd)
// 		case "SYNC_RAW_END":
// 			slavelog.Printf("[M %s] recv bulk finish\n", s.session.RemoteAddr())
// 		case "SYNC_SEQ_START":
// 			slavelog.Printf("[M %s] sync online ...\n", s.session.RemoteAddr())
// 			s.Session().SetAttribute(S_STATUS, REPL_ONLINE)
// 			s.recvCommandSeq(cmd) // 进入后只有出错才退出
// 			break
// 		default:
// 			s.server.On(s.session, cmd)
// 		}
// 	}
// 	return
// }

// func (s *SlaveClientV2) Close() {
// 	s.session.Close()
// 	s.synclog.Close()
// 	return
// }

// // 收取快照完成后，开始收取实时数据
// func (s *SlaveClientV2) recvCommandSeq(cmd *Command) (err error) {
// 	session := s.session
// 	for {
// 		// SYNC_SEQ
// 		cmd, err = session.ReadCommand()
// 		if err != nil {
// 			break
// 		}
// 		cmdName := cmd.Name()
// 		switch cmdName {
// 		case "PING":
// 			continue
// 		case "SYNC_SEQ":
// 			s.lastseq, err = cmd.Int64AtIndex(1)
// 			if err != nil {
// 				break
// 			}
// 		default: // commands
// 			s.counters.Get("cmd").Incr(1)
// 			s.server.On(session, cmd)
// 		}
// 	}
// 	return
// }

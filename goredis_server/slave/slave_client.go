package slave

import (
	. "GoRedis/goredis"
	"GoRedis/libs/statlog"
	"io"
	"os"
)

type SlaveClient struct {
	session *Session
	home    string
	buffer  chan *Command
	synclog *statlog.StatLogger
}

func NewClient(session *Session, home string) (s *SlaveClient, err error) {
	s = &SlaveClient{}
	s.home = home
	s.session = session
	s.buffer = make(chan *Command, 1000*10000)

	if err = s.initSyncLog(home + "/sync.log"); err != nil {
		return
	}
	return
}

func (s *SlaveClient) initSyncLog(filename string) error {
	if file, err := os.OpenFile(filename, os.O_CREATE|os.O_APPEND|os.O_WRONLY, os.ModePerm); err != nil {
		return err
	} else {
		s.synclog = statlog.NewStatLogger(file)
	}
	s.synclog.Add(statlog.TimeItem("time"))
	s.synclog.Add(statlog.Item("send", func() interface{} {
		return nil
	}, &statlog.Opt{Padding: 10}))
	s.synclog.Add(statlog.Item("recv", func() interface{} {
		return nil
	}, &statlog.Opt{Padding: 10}))
	s.synclog.Add(statlog.Item("buffer", func() interface{} {
		return nil
	}, &statlog.Opt{Padding: 10}))
	return nil
}

func (s *SlaveClient) SendRdb(r io.Reader) (err error) {
	return
}

func (s *SlaveClient) SendCommand(cmd *Command) (err error) {
	s.buffer <- cmd
	return
}

func (s *SlaveClient) Sync(uid string) (err error) {

	return
}

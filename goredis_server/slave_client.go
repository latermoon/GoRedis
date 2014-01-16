package goredis_server

import (
	. "../goredis"
	"../libs/iotool"
	"../libs/stdlog"
	"bufio"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"strings"
)

type SlaveStatus int

type SlaveClientCallback interface {
	RdbSizeCallback(size int64)
	RdbRecvCallback(r *bufio.Reader)
	IdleCallback()
	CommandRecvCallback(cmd *Command)
}

/**

client := NewSlaveClient(...)
client.Sync(uid)
client.Cancel()

*/
type SlaveClient struct {
	session  *Session
	callback SlaveClientCallback
}

func NewSlaveClient(session *Session) (s *SlaveClient) {
	s = &SlaveClient{}
	s.session = session
	return
}

func (s *SlaveClient) SetCallback(callback SlaveClientCallback) {
	s.callback = callback
}

// 开始同步
func (s *SlaveClient) Sync(uid string) (err error) {
	isgoredis, version, e1 := s.masterInfo()
	if e1 != nil {
		return e1
	}
	if isgoredis {
		stdlog.Println("slaveof GoRedis:", version)
	} else {
		stdlog.Println("slaveof Redis:", version)
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
		if !rdbsaved && c == '$' {
			err = s.recvRdb()
			if err != nil {
				stdlog.Println("recv rdb error:", err)
				break
			}
			rdbsaved = true
		} else if c == '\n' {
			s.session.ReadByte()
			s.callback.IdleCallback()
		} else {
			var cmd *Command
			cmd, err = s.session.ReadCommand()
			if err != nil {
				break
			}
			s.callback.CommandRecvCallback(cmd)
		}
	}
	return
}

func (s *SlaveClient) recvRdb() (err error) {
	var f *os.File
	f, err = ioutil.TempFile("/tmp/", "tmp_goredis_")
	if err != nil {
		return
	}
	defer func() {
		filename := f.Name()
		f.Close()
		os.Remove(filename)
	}()

	s.session.ReadByte()
	var size int64
	size, err = s.session.ReadLineInteger()
	if err != nil {
		return
	}
	s.callback.RdbSizeCallback(size)

	// read
	w := bufio.NewWriter(f)
	_, err = iotool.RateLimitCopy(w, io.LimitReader(s.session, size), 200*1024*1024, func(written int64, rate int) {
		stdlog.Println("copy:", written, "rate:", rate)
	})
	// _, err = io.CopyN(w, s.session, size)
	if err != nil {
		return
	}
	w.Flush()

	// callback
	s.callback.RdbRecvCallback(bufio.NewReader(f))
	return
}

// 清空本地的同步状态
func (s *SlaveClient) Destory() (err error) {
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

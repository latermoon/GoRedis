package goredis_server

import (
	. "../goredis"
	"errors"
	"strings"
)

type SlaveStatus int

/**

client := NewSlaveClient(...)
client.Sync(uid)
client.Cancel()

*/
type SlaveClient struct {
	session *Session
}

func NewSlaveClient(session *Session) (s *SlaveClient) {
	s = &SlaveClient{}
	s.session = session
	return
}

// 开始同步
func (s *SlaveClient) Sync(uid string) (err error) {
	return
}

// 清空本地的同步状态
func (s *SlaveClient) Destory() (err error) {
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

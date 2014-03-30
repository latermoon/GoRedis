package goredis_server

// 管理发出monitor指令的连接，传输实时指令
import (
	. "GoRedis/goredis"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"sync"
	"time"
)

type MonClient struct {
	session *Session
	buffer  chan string
	closed  bool
	mu      sync.Mutex
}

func NewMonClient(session *Session) (m *MonClient) {
	m = &MonClient{
		session: session,
		buffer:  make(chan string, 10000),
	}
	return
}

func (m *MonClient) Send(cmdex *CommandEx) (err error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	if m.closed {
		return errors.New("buffer closed")
	}
	if len(m.buffer) == cap(m.buffer) {
		m.Close()
		return errors.New("out of buffer limit")
	}

	line := m.formatCommandLine(cmdex)
	m.buffer <- line
	return
}

func (m *MonClient) Start() (err error) {
	for {
		line, ok := <-m.buffer
		if !ok {
			err = errors.New("buffer broken")
			break
		}
		err = m.session.WriteReply(StatusReply(line))
		if err != nil {
			break
		}
	}
	m.Close()
	return
}

func (m *MonClient) Close() {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.closed = true

	close(m.buffer)
	m.session.Close()
}

// 将Command转换为下面格式
// +1386347668.732167 [0 10.80.101.169:8400] "ZADD" "user:update:timestamp" "1.386347668E9" "40530990"
func (m *MonClient) formatCommandLine(cmdex *CommandEx) (s string) {
	// 对于cmd，用json编码，然后去掉前后的"[]"以及中间的逗号"," ["SET", "name", "latermoon"] => "SET" "name" "lateroon"
	args := make([]string, cmdex.Len())
	for i, b := range cmdex.Args {
		args[i] = string(b)
	}
	b, err := json.Marshal(args)
	cmdstr := string(b)
	if err != nil {
		cmdstr = cmdex.String()
	} else if len(cmdstr) >= 2 {
		cmdstr = cmdstr[1 : len(cmdstr)-1] // trim "[" & "]"
		cmdstr = strings.Replace(cmdstr, "\",\"", "\" \"", -1)
	}
	s = fmt.Sprintf("+%f [0 %s] %s", float64(time.Now().UnixNano())/1e9, cmdex.Session().RemoteAddr(), cmdstr)
	return
}

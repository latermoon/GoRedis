package goredis_server

// 管理当前发起monitor的连接
import (
	. "GoRedis/goredis"
	"container/list"
	"encoding/json"
	"fmt"
	"strings"
	"sync"
	"time"
)

type MonManager struct {
	clients *list.List
	mu      sync.Mutex
}

func NewMonManager() (m *MonManager) {
	m = &MonManager{
		clients: list.New(),
	}
	return
}

func (m *MonManager) Count() int {
	return m.clients.Len()
}

// 广播同步
func (m *MonManager) BroadcastCommand(cmd *Command) {
	if m.clients.Len() == 0 {
		return
	}
	template := monitorLineTemplate(cmd)
	m.mu.Lock()
	defer m.mu.Unlock()
	for e := m.clients.Front(); e != nil; e = e.Next() {
		c := e.Value.(*MonClient)
		line := strings.Replace(template, "%s", c.session.RemoteAddr().String(), 1)
		err := c.Send(line)
		if err != nil {
			errlog.Println("[%s] monitor broadcast error %s", c.session.RemoteAddr(), err)
			m.clients.Remove(e)
		}
	}
}

// 格式化后留空session addr，用于多monitor时减少格式化成本
// +1386347668.732167 [0 10.80.101.169:8400] "ZADD" "user:update:timestamp" "1.386347668E9" "40530990"
func monitorLineTemplate(cmd *Command) (s string) {
	// 对于cmd，用json编码，然后去掉前后的"[]"以及中间的逗号","
	// ["SET", "name", "latermoon"] => "SET" "name" "lateroon"
	b, err := json.Marshal(cmd.StringArgs())
	cmdstr := string(b)
	if err != nil {
		cmdstr = cmd.String()
	} else if len(cmdstr) >= 2 {
		cmdstr = cmdstr[1 : len(cmdstr)-1] // trim "[" & "]"
		cmdstr = strings.Replace(cmdstr, "\",\"", "\" \"", -1)
	}
	s = fmt.Sprintf("+%f [0 %s] %s", float64(time.Now().UnixNano())/1e9, "%s", cmdstr)
	return
}

func (m *MonManager) Add(c *MonClient) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.clients.PushBack(c)
}

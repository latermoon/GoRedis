package goredis_server

import (
	. "../goredis"
	"container/list"
	"encoding/json"
	"fmt"
	"strings"
	"time"
)

// 在On(*Session, *Command)里使用goroutine触发
func (server *GoRedisServer) monitorOutput(session *Session, cmd *Command) {
	server.monitorMutex.Lock()
	defer server.monitorMutex.Unlock()

	if server.monitorlist.Len() == 0 {
		return
	}
	// 存放要移出的session
	var needRemove []*list.Element
	for e := server.monitorlist.Front(); e != nil; e = e.Next() {
		s := e.Value.(*Session)

		line := server.formatMonitorCommand(session, cmd)
		err := s.WriteReply(StatusReply(line))
		if err != nil {
			if needRemove == nil {
				needRemove = make([]*list.Element, 0, 1)
			}
			needRemove = append(needRemove, e)
		}
	}
	if needRemove != nil {
		for _, e := range needRemove {
			stdlog.Info("remove monitor client %s", e.Value.(*Session).RemoteAddr())
			server.monitorlist.Remove(e)
		}
	}
}

// +1386347668.732167 [0 10.80.101.169:8400] "ZADD" "user:update:timestamp" "1.386347668E9" "40530990"
func (server *GoRedisServer) formatMonitorCommand(session *Session, cmd *Command) (s string) {
	t := time.Now()
	// 对于cmd，用json编码，然后去掉前后的"[]"以及中间的逗号","
	// ["SET", "name", "latermoon"] => "SET" "name" "lateroon"
	b, err := json.Marshal(cmd.StringArgs())
	cmdstr := string(b)
	if err != nil {
		cmdstr = cmd.String()
	} else {
		cmdstr = strings.TrimPrefix(cmdstr, "[")
		cmdstr = strings.Replace(cmdstr, "\",\"", "\" \"", -1)
		cmdstr = strings.TrimSuffix(cmdstr, "]")
	}
	s = fmt.Sprintf("+%f [0 %s] %s", float64(t.UnixNano())/1e9, session.RemoteAddr(), cmdstr)
	return
}

func (server *GoRedisServer) OnMONITOR(session *Session, cmd *Command) (reply *Reply) {
	session.WriteReply(StatusReply("OK"))

	server.monitorMutex.Lock()
	defer server.monitorMutex.Unlock()
	server.monitorlist.PushBack(session)
	return
}

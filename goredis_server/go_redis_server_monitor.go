package goredis_server

import (
	. "GoRedis/libs/goredis"
	"GoRedis/libs/stdlog"
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
			stdlog.Printf("remove monitor client %s\n", e.Value.(*Session).RemoteAddr())
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

// redis-cli对monitor指令进行特殊处理，只要monitor不断输出StatusReply，可以实现不间断的流输出
// 适用于海量数据的扫描输出，比如iterator扫描整个数据库
func (server *GoRedisServer) OnMONITOR(session *Session, cmd *Command) (reply *Reply) {
	// 特殊使用，monitor输出全部key
	if len(cmd.Args) > 1 && strings.ToLower(cmd.StringAtIndex(1)) == "keys" {
		server.monitorKeys(session, cmd)
		return
	}

	session.WriteReply(StatusReply("OK"))
	server.monitorMutex.Lock()
	defer server.monitorMutex.Unlock()
	server.monitorlist.PushBack(session)
	return
}

// echo 'monitor keys' | redis-cli -p 1602
func (server *GoRedisServer) monitorKeys(session *Session, cmd *Command) {
	prefix, _ := cmd.ArgAtIndex(2)
	server.levelRedis.Keys(prefix, func(i int, key, keytype []byte, quit *bool) {
		err := session.WriteReply(StatusReply(string(key)))
		if err != nil {
			*quit = true
		}
	})
	session.Close()
}

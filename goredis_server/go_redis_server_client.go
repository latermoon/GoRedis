package goredis_server

import (
	. "GoRedis/goredis"
	"bytes"
	"fmt"
	"strings"
)

func (server *GoRedisServer) OnCLIENT(session *Session, cmd *Command) (reply *Reply) {
	switch strings.ToUpper(cmd.StringAtIndex(1)) {
	case "LIST":
		reply = server.replyClientList(session, cmd)
	default:
		reply = ErrorReply("not support")
	}
	return
}

// Get the list of client connections
func (server *GoRedisServer) replyClientList(session *Session, cmd *Command) (reply *Reply) {
	buf := bytes.Buffer{}
	server.sessmgr.Enumerate(func(i int, key string, val interface{}) {
		sess := val.(*Session)
		lastcmd := sess.GetAttribute("cmd")
		if lastcmd == nil {
			lastcmd = ""
		}
		buf.WriteString(fmt.Sprintf("addr=%s i=%d cmd=%s\n", key, i, lastcmd))
	})
	reply = BulkReply(buf.Bytes())
	return
}

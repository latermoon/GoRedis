package goredis_server

import (
	. "../goredis"
	"net"
)

// 临时命名，实现功能后重构
type SessionEx struct {
	Session
	conn net.Conn
}

func newSessionEx(s *Session) (sess *SessionEx) {
	sess = &SessionEx{}
	sess.conn = s.Connection()
	return
}

/*
*<number of arguments> CR LF
$<number of bytes of argument 1> CR LF
<argument data> CR LF
...
$<number of bytes of argument N> CR LF
<argument data> CR LF
*/
func (s *SessionEx) Send(cmd *Command) (err error) {
	_, err = s.conn.Write(cmd.Bytes())
	return
}

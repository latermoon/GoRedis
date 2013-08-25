// http://redis.io/topics/protocol
package goredis

import (
	"bytes"
	"net"
	"strconv"
)

// ==============================
// Session维护一个net.Conn连接，代表一个客户端会话
// 提供各种标准的Reply方法
// ==============================
type Session struct {
	conn net.Conn
}

func newSession(conn net.Conn) (session *Session) {
	session = &Session{conn: conn}
	return
}

// Status reply
func (session *Session) ReplyStatus(status string) (err error) {
	err = nil
	buf := bytes.Buffer{}
	buf.WriteString("+")
	buf.WriteString(status)
	buf.WriteString(CRLF)
	buf.WriteTo(session.conn)
	return
}

// Error reply
func (session *Session) ReplyError(errmsg string) (err error) {
	err = nil
	buf := bytes.Buffer{}
	buf.WriteString("-")
	buf.WriteString(errmsg)
	buf.WriteString(CRLF)
	buf.WriteTo(session.conn)
	return
}

// Integer reply
func (session *Session) ReplyInteger(i int) (err error) {
	err = nil
	buf := bytes.Buffer{}
	buf.WriteString(":")
	buf.WriteString(strconv.Itoa(i))
	buf.WriteString(CRLF)
	buf.WriteTo(session.conn)
	return
}

// Bulk Reply
func (session *Session) ReplyBulk(bulk interface{}) (err error) {
	err = nil
	// NULL Bulk Reply
	if bulk == nil {
		session.conn.Write([]byte("$-1\r\n"))
		return
	}
	buf := bytes.Buffer{}
	buf.WriteString("$")
	switch bulk.(type) {
	case []byte:
		b := bulk.([]byte)
		buf.WriteString(strconv.Itoa(len(b)))
		buf.WriteString(CRLF)
		buf.Write(b)
	default:
		b := []byte(bulk.(string))
		buf.WriteString(strconv.Itoa(len(b)))
		buf.WriteString(CRLF)
		buf.Write(b)
	}
	buf.WriteString(CRLF)
	buf.WriteTo(session.conn)
	return
}

// Multi-bulk replies
func (session *Session) ReplyMultiBulks(bulks []interface{}) (err error) {
	err = nil
	// Null Multi Bulk Reply
	if bulks == nil {
		session.conn.Write([]byte("*-1\r\n"))
		return
	}
	bulkCount := len(bulks)
	// Empty Multi Bulk Reply
	if bulkCount == 0 {
		session.conn.Write([]byte("*0\r\n"))
		return
	}
	buf := bytes.Buffer{}
	buf.WriteString("*")
	buf.WriteString(strconv.Itoa(bulkCount))
	buf.WriteString(CRLF)
	for i := 0; i < bulkCount; i++ {
		bulk := bulks[i]
		switch bulk.(type) {
		case string:
			buf.WriteString("$")
			b := []byte(bulk.(string))
			buf.WriteString(strconv.Itoa(len(b)))
			buf.WriteString(CRLF)
			buf.Write(b)
			buf.WriteString(CRLF)
		case []byte:
			buf.WriteString("$")
			b := bulk.([]byte)
			buf.WriteString(strconv.Itoa(len(b)))
			buf.WriteString(CRLF)
			buf.Write(b)
			buf.WriteString(CRLF)
		case int:
			buf.WriteString(":")
			buf.WriteString(strconv.Itoa(bulk.(int)))
			buf.WriteString(CRLF)
		default:
			// nil element
			buf.WriteString("$-1")
			buf.WriteString(CRLF)
		}
	}
	// flush
	buf.WriteTo(session.conn)
	return
}

// Close conn
func (session *Session) Close() error {
	return session.conn.Close()
}

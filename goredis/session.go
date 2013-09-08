package goredis

import (
	"bufio"
	"bytes"
	"errors"
	"fmt"
	"net"
	"strconv"
)

// ==============================
// Session维护一个net.Conn连接，代表一个客户端会话
// 提供各种标准的Reply方法, Status/Error/Integer/Bulk/MultiBulks
// cmd := session.ReadCommand()
// session.Reply(IntegerReply(10))
// session.Reply(StatusReply("OK"))
// 协议参考：http://redis.io/topics/protocol
// ==============================
type Session struct {
	conn   net.Conn
	reader *bufio.Reader
}

// 创建Session
func newSession(conn net.Conn) (s *Session) {
	s = &Session{}
	s.conn = conn
	s.reader = bufio.NewReader(s.conn)
	return
}

// 返回数据到客户端
func (s *Session) Reply(reply *Reply) (err error) {
	switch reply.Type {
	case ReplyTypeStatus:
		err = s.replyStatus(reply.Value.(string))
	case ReplyTypeError:
		err = s.replyError(reply.Value.(string))
	case ReplyTypeInteger:
		err = s.replyInteger(reply.Value.(int))
	case ReplyTypeBulk:
		err = s.replyBulk(reply.Value)
	case ReplyTypeMultiBulks:
		err = s.replyMultiBulks(reply.Value.([]interface{}))
	default:
		err = errors.New("Illegal ReplyType: " + strconv.Itoa(int(reply.Type)))
		panic(err)
	}
	return
}

// Status reply
func (s *Session) replyStatus(status string) (err error) {
	buf := bytes.Buffer{}
	buf.WriteString("+")
	buf.WriteString(status)
	buf.WriteString(CRLF)
	buf.WriteTo(s.conn)
	return
}

// Error reply
func (s *Session) replyError(errmsg string) (err error) {
	buf := bytes.Buffer{}
	buf.WriteString("-")
	buf.WriteString(errmsg)
	buf.WriteString(CRLF)
	buf.WriteTo(s.conn)
	return
}

// Integer reply
func (s *Session) replyInteger(i int) (err error) {
	buf := bytes.Buffer{}
	buf.WriteString(":")
	buf.WriteString(strconv.Itoa(i))
	buf.WriteString(CRLF)
	buf.WriteTo(s.conn)
	return
}

// Bulk Reply
func (s *Session) replyBulk(bulk interface{}) (err error) {
	// NULL Bulk Reply
	if bulk == nil {
		s.conn.Write([]byte("$-1\r\n"))
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
	buf.WriteTo(s.conn)
	return
}

// Multi-bulk replies
func (s *Session) replyMultiBulks(bulks []interface{}) (err error) {
	// Null Multi Bulk Reply
	if bulks == nil {
		s.conn.Write([]byte("*-1\r\n"))
		return
	}
	bulkCount := len(bulks)
	// Empty Multi Bulk Reply
	if bulkCount == 0 {
		s.conn.Write([]byte("*0\r\n"))
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
	buf.WriteTo(s.conn)
	return
}

// Close conn
func (s *Session) Close() error {
	return s.conn.Close()
}

func (s *Session) String() string {
	return fmt.Sprintf("<Session:%s>", s.conn.RemoteAddr().String())
}

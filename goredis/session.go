// http://redis.io/topics/protocol
package goredis

import (
	"bufio"
	"bytes"
	"errors"
	"net"
	"strconv"
)

// ==============================
// Session维护一个net.Conn连接，代表一个客户端会话
// 提供各种标准的Reply方法, Status/Error/Integer/Bulk/MultiBulk
// 在性能损耗允许范围内，可以整合成一个接口
// session.Reply(IntegerReply(10))
// session.Reply(StatusReply("OK"))
// ==============================
type Session struct {
	conn   net.Conn
	reader *bufio.Reader
}

func newSession(conn net.Conn) (s *Session) {
	s = &Session{}
	s.conn = conn
	s.reader = bufio.NewReader(s.conn)
	return
}

// 从客户端连接获取指令
// (下面读取过程，线上应用前需要增加错误校验，数据大小限制)
/*
*<number of arguments> CR LF
$<number of bytes of argument 1> CR LF
<argument data> CR LF
...
$<number of bytes of argument N> CR LF
<argument data> CR LF
*/
func (s *Session) readCommand() (cmd *Command, err error) {
	reader := s.reader

	cmd = &Command{}
	err = nil

	// Read ( *<number of arguments> CR LF )
	_, err = reader.ReadBytes('*')
	if err != nil { // EOF
		return
	}
	line, e1 := reader.ReadBytes(CR)
	if e1 != nil {
		err = e1
		return
	}
	// number of arguments
	argCount, _ := strconv.Atoi(string(line[:len(line)-1]))
	_, err = reader.ReadBytes(LF)
	if err != nil {
		return
	}

	cmd.Args = make([][]byte, argCount)
	for i := 0; i < argCount; i++ {
		// Read ( $<number of bytes of argument 1> CR LF )
		_, _ = reader.ReadBytes('$')
		line, e2 := reader.ReadBytes(CR)
		if e2 != nil {
			err = e2
			return
		}
		argSize, _ := strconv.Atoi(string(line[:len(line)-1]))
		_, err = reader.ReadBytes(LF)

		// Read ( <argument data> CR LF )
		cmd.Args[i] = make([]byte, argSize)
		_, err = reader.Read(cmd.Args[i])
		_, err = reader.ReadBytes(CR)
		_, err = reader.ReadBytes(LF)
		if err != nil {
			return
		}
	}

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
		err = errors.New("Bad ReplyType")
		panic(err)
	}
	return
}

// Status reply
func (s *Session) replyStatus(status string) (err error) {
	err = nil
	buf := bytes.Buffer{}
	buf.WriteString("+")
	buf.WriteString(status)
	buf.WriteString(CRLF)
	buf.WriteTo(s.conn)
	return
}

// Error reply
func (s *Session) replyError(errmsg string) (err error) {
	err = nil
	buf := bytes.Buffer{}
	buf.WriteString("-")
	buf.WriteString(errmsg)
	buf.WriteString(CRLF)
	buf.WriteTo(s.conn)
	return
}

// Integer reply
func (s *Session) replyInteger(i int) (err error) {
	err = nil
	buf := bytes.Buffer{}
	buf.WriteString(":")
	buf.WriteString(strconv.Itoa(i))
	buf.WriteString(CRLF)
	buf.WriteTo(s.conn)
	return
}

// Bulk Reply
func (s *Session) replyBulk(bulk interface{}) (err error) {
	err = nil
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
	err = nil
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

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
func (s *Session) ReadCommand() (cmd *Command, err error) {
	reader := s.reader

	cmd = &Command{}
	err = nil
	var c byte
	var line []byte
	// Read ( *<number of arguments> CR LF )
	if c, err = reader.ReadByte(); err != nil { // io.EOF
		return
	} else if c != '*' {
		err = errors.New("Illegal * ...")
		return
	}
	if line, err = lightReadBytes(reader, CR); err != nil {
		return
	}
	// number of arguments
	argCount, _ := strconv.Atoi(string(line))
	if c, err = reader.ReadByte(); err != nil {
		return
	} else if c != LF {
		err = errors.New("Illegal LF 1 ...")
		return
	}

	cmd.Args = make([][]byte, argCount)
	for i := 0; i < argCount; i++ {
		// Read ( $<number of bytes of argument 1> CR LF )
		if c, err = reader.ReadByte(); err != nil {
			return
		} else if c != '$' {
			err = errors.New("Illegal $ ...")
			return
		}

		if line, err = lightReadBytes(reader, CR); err != nil {
			return
		}
		argSize, _ := strconv.Atoi(string(line))
		if c, err = reader.ReadByte(); err != nil {
			return
		} else if c != LF {
			err = errors.New("Illegal LF 2 ...")
			return
		}

		// Read ( <argument data> CR LF )
		cmd.Args[i] = make([]byte, argSize)
		// 这里要注意是否填充完整
		var n int
		if n, err = reader.Read(cmd.Args[i]); err != nil {
			return
		} else if n != argSize {
			err = errors.New("Broken Pipe")
		}

		if c, err = reader.ReadByte(); err != nil {
			return
		} else if c != CR {
			err = errors.New("Illegal CR ...")
			return
		}

		if c, err = reader.ReadByte(); err != nil {
			return
		} else if c != LF {
			err = errors.New("Illegal LF 3 ...")
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
		err = errors.New("Illegal ReplyType: " + strconv.Itoa(int(reply.Type)))
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

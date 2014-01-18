// Copyright (c) 2013, Latermoon <lptmoon@gmail.com>
// All rights reserved.
//
package goredis

import (
	"bufio"
	"bytes"
	"errors"
	"fmt"
	"io"
	"net"
	"strconv"
)

// ==============================
// Session维护一个net.Conn连接，代表一个客户端会话
// 提供各种标准的Reply方法, Status/Error/Integer/Bulk/MultiBulks
// cmd, err := session.ReadCommand()
// session.Reply(StatusReply("OK"))
// 协议参考：http://redis.io/topics/protocol
// ==============================
type Session struct {
	*bufio.ReadWriter // 实现了Read和Write方法即可
	conn              net.Conn
	rw                *bufio.ReadWriter
}

func NewSession(conn net.Conn) (s *Session) {
	s = &Session{}
	s.conn = conn
	s.rw = bufio.NewReadWriter(bufio.NewReader(conn), bufio.NewWriter(conn))
	return
}

func (s *Session) LocalAddr() net.Addr {
	return s.conn.LocalAddr()
}

func (s *Session) RemoteAddr() net.Addr {
	return s.conn.RemoteAddr()
}

// 返回数据到客户端
func (s *Session) Reply(reply *Reply) (err error) {
	return s.WriteReply(reply)
}

func (s *Session) WriteReply(reply *Reply) (err error) {
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
	}
	return
}

func (s *Session) WriteCommand(cmd *Command) (err error) {
	_, err = s.rw.Write(cmd.Bytes())
	if err == nil {
		err = s.rw.Flush()
	}
	return
}

// 从连接里读取回复
/*
In a Status Reply the first byte of the reply is "+"
In an Error Reply the first byte of the reply is "-"
In an Integer Reply the first byte of the reply is ":"
In a Bulk Reply the first byte of the reply is "$"
In a Multi Bulk Reply the first byte of the reply s "*"
*/
func (s *Session) ReadReply() (reply *Reply, err error) {
	reader := s.rw
	var c byte
	if c, err = reader.ReadByte(); err != nil {
		return
	}

	reply = &Reply{}
	switch c {
	case '+':
		reply.Type = ReplyTypeStatus
		reply.Value, err = s.readString()
	case '-':
		reply.Type = ReplyTypeError
		reply.Value, err = s.readString()
	case ':':
		reply.Type = ReplyTypeInteger
		reply.Value, err = s.readInt()
	case '$':
		reply.Type = ReplyTypeBulk
		var bufsize int
		bufsize, err = s.readInt()
		if err != nil {
			break
		}
		buf := make([]byte, bufsize)
		_, err = io.ReadFull(s, buf)
		if err != nil {
			break
		}
		reply.Value = buf
		s.skipBytes([]byte{CR, LF})
	case '*':
		reply.Type = ReplyTypeMultiBulks
		var argCount int
		argCount, err = s.readInt()
		if err != nil {
			break
		}
		if argCount == -1 {
			reply.Value = nil // *-1
		} else {
			args := make([][]byte, argCount)
			for i := 0; i < argCount; i++ {
				// TODO multi bulk 的类型 $和:
				err = s.skipByte('$')
				if err != nil {
					break
				}
				var argSize int
				argSize, err = s.readInt()
				if err != nil {
					return
				}
				if argSize == -1 {
					args[i] = nil
				} else {
					args[i] = make([]byte, argSize)
					_, err = io.ReadFull(s, args[i])
					if err != nil {
						break
					}
				}
				s.skipBytes([]byte{CR, LF})
			}
			reply.Value = args
		}
	default:
		err = errors.New("Bad Reply Flag:" + string([]byte{c}))
	}
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
	cmd = &Command{}

	// Read ( *<number of arguments> CR LF )
	err = s.skipByte('*')
	if err != nil { // io.EOF
		return
	}
	// number of arguments
	var argCount int
	if argCount, err = s.readInt(); err != nil {
		return
	}
	cmd.Args = make([][]byte, argCount)
	for i := 0; i < argCount; i++ {
		// Read ( $<number of bytes of argument 1> CR LF )
		err = s.skipByte('$')
		if err != nil {
			return
		}

		var argSize int
		argSize, err = s.readInt()
		if err != nil {
			return
		}

		// Read ( <argument data> CR LF )
		cmd.Args[i] = make([]byte, argSize)
		_, err = io.ReadFull(s, cmd.Args[i])
		if err != nil {
			return
		}

		err = s.skipBytes([]byte{CR, LF})
		if err != nil {
			return
		}
	}

	return
}

// Status reply
func (s *Session) replyStatus(status string) (err error) {
	buf := bytes.Buffer{}
	buf.WriteString("+")
	buf.WriteString(status)
	buf.WriteString(CRLF)
	_, err = buf.WriteTo(s.conn)
	return
}

// Error reply
func (s *Session) replyError(errmsg string) (err error) {
	buf := bytes.Buffer{}
	buf.WriteString("-")
	buf.WriteString(errmsg)
	buf.WriteString(CRLF)
	_, err = buf.WriteTo(s.conn)
	return
}

// Integer reply
func (s *Session) replyInteger(i int) (err error) {
	buf := bytes.Buffer{}
	buf.WriteString(":")
	buf.WriteString(strconv.Itoa(i))
	buf.WriteString(CRLF)
	_, err = buf.WriteTo(s.conn)
	return
}

// Bulk Reply
func (s *Session) replyBulk(bulk interface{}) (err error) {
	// NULL Bulk Reply
	isnil := bulk == nil
	if !isnil {
		// []byte 需要类型转换后才能判断
		b, ok := bulk.([]byte)
		isnil = ok && b == nil
	}
	if isnil {
		_, err = s.conn.Write([]byte("$-1\r\n"))
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
	_, err = buf.WriteTo(s.conn)
	return
}

// Multi-bulk replies
func (s *Session) replyMultiBulks(bulks []interface{}) (err error) {
	// Null Multi Bulk Reply
	if bulks == nil {
		_, err = s.conn.Write([]byte("*-1\r\n"))
		return
	}
	bulkCount := len(bulks)
	// Empty Multi Bulk Reply
	if bulkCount == 0 {
		_, err = s.conn.Write([]byte("*0\r\n"))
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
			b := bulk.([]byte)
			if b == nil {
				buf.WriteString("$-1")
				buf.WriteString(CRLF)
			} else {
				buf.WriteString("$")
				buf.WriteString(strconv.Itoa(len(b)))
				buf.WriteString(CRLF)
				buf.Write(b)
				buf.WriteString(CRLF)
			}
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
	_, err = buf.WriteTo(s.conn)
	return
}

// ====================================
// io
// ====================================

// 验证并跳过指定的字节，用于开始符和结束符的判断
func (s *Session) skipByte(c byte) (err error) {
	var tmp byte
	tmp, err = s.rw.ReadByte()
	if err != nil {
		return
	}
	if tmp != c {
		err = errors.New(fmt.Sprintf("Illegal Byte [%d] != [%d]", tmp, c))
	}
	return
}

func (s *Session) skipBytes(bs []byte) (err error) {
	for _, c := range bs {
		err = s.skipByte(c)
		if err != nil {
			break
		}
	}
	return
}

// 读取一行
func (s *Session) readLine() (line []byte, err error) {
	line, err = s.rw.ReadSlice(LF)
	if err == bufio.ErrBufferFull {
		return nil, errors.New("line too long")
	}
	if err != nil {
		return
	}
	i := len(line) - 2
	if i < 0 || line[i] != CR {
		err = errors.New("bad line terminator:" + string(line))
	}
	return line[:i], nil
}

// 读取字符串，遇到CRLF换行为止
func (s *Session) readString() (str string, err error) {
	var line []byte
	if line, err = s.readLine(); err != nil {
		return
	}
	str = string(line)
	return
}

func (s *Session) readInt() (i int, err error) {
	var line string
	if line, err = s.readString(); err != nil {
		return
	}
	i, err = strconv.Atoi(line)
	return
}

func (s *Session) readInt64() (i int64, err error) {
	var line string
	if line, err = s.readString(); err != nil {
		return
	}
	i, err = strconv.ParseInt(line, 10, 64)
	return
}

func (s *Session) ReadInt64() (i int64, err error) {
	return s.readInt64()
}

func (s *Session) Read(p []byte) (n int, err error) {
	return s.rw.Read(p)
}

func (s *Session) Write(p []byte) (n int, err error) {
	n, err = s.rw.Write(p)
	if err == nil {
		s.rw.Flush()
	}
	return
}

// Close conn
func (s *Session) Close() error {
	return s.conn.Close()
}

func (s *Session) ReadByte() (c byte, err error) {
	return s.rw.ReadByte()
}

func (s *Session) ReadBytes(delim byte) (line []byte, err error) {
	return s.rw.ReadBytes(delim)
}

// 获取字节而不移动游标
func (s *Session) PeekByte() (c byte, err error) {
	c, err = s.rw.ReadByte()
	if err == nil {
		err = s.rw.UnreadByte()
	}
	return
}

func (s *Session) ReadRDB(w io.Writer) (err error) {
	// Read ( $<number of bytes of RDB> CR LF )
	if err = s.skipByte('$'); err != nil {
		return
	}

	var rdbSize int64
	if rdbSize, err = s.readInt64(); err != nil {
		return
	}

	var c byte
	for i := int64(0); i < rdbSize; i++ {
		c, err = s.rw.ReadByte()
		if err != nil {
			return
		}
		w.Write([]byte{c})
	}
	return
}

func (s *Session) String() string {
	return fmt.Sprintf("<Session:%s>", s.RemoteAddr())
}

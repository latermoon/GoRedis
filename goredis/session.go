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
// session.WriteReply(IntegerReply(10))
// session.WriteReply(StatusReply("OK"))
// 协议参考：http://redis.io/topics/protocol
// ==============================
type Session struct {
	*bufio.ReadWriter // 实现了Read和Write方法即可
	conn              net.Conn
	rw                *bufio.ReadWriter
	curCmd            *Command // 每次ReadCommand使用同一个对象
}

func NewSession(conn net.Conn) (s *Session) {
	s = &Session{}
	s.conn = conn
	s.rw = bufio.NewReadWriter(bufio.NewReader(conn), bufio.NewWriter(conn))
	s.curCmd = &Command{}
	s.curCmd.Args = make([][]byte, 0, 10)
	return
}

func (s *Session) LocalAddr() net.Addr {
	return s.conn.LocalAddr()
}

func (s *Session) RemoteAddr() net.Addr {
	return s.conn.RemoteAddr()
}

func (s *Session) reuseCommand(argCount int) (cmd *Command) {
	if argCount > cap(s.curCmd.Args) {
		// 扩容
		s.curCmd.Args = make([][]byte, 0, argCount)
	} else {
		// 截断复用
		s.curCmd.Args = s.curCmd.Args[0:0]
	}
	cmd = s.curCmd
	return
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
	reply = &Reply{}
	var c byte
	if c, err = reader.ReadByte(); err != nil {
		return
	}

	switch c {
	case '+':
		reply.Type = ReplyTypeStatus
		reply.Value, err = s.readLineString()
	case '-':
		reply.Type = ReplyTypeError
		reply.Value, err = s.readLineString()
	case ':':
		reply.Type = ReplyTypeInteger
		reply.Value, err = s.readLineInteger()
	case '$':
		reply.Type = ReplyTypeBulk
		var bufsize int
		bufsize, err = s.readLineInteger()
		if err != nil {
			break
		}
		reply.Value, err = s.BlockReadBytes(bufsize)
		if err != nil {
			break
		}
		s.skipSpecificBytes([]byte{CR, LF})
	case '*':
		reply.Type = ReplyTypeMultiBulks
		var argCount int
		argCount, err = s.readLineInteger()
		if err != nil {
			break
		}
		if argCount == -1 {
			reply.Value = nil // *-1
		} else {
			args := make([][]byte, argCount)
			for i := 0; i < argCount; i++ {
				// TODO multi bulk 的类型 $和:
				err = s.skipSpecificByte('$')
				if err != nil {
					break
				}
				var argSize int
				argSize, err = s.readLineInteger()
				if err != nil {
					return
				}
				if argSize == -1 {
					args[i] = nil
				} else {
					args[i], err = s.BlockReadBytes(argSize)
					if err != nil {
						break
					}
				}
				s.skipSpecificBytes([]byte{CR, LF})
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
	// Read ( *<number of arguments> CR LF )
	err = s.skipSpecificByte('*')
	if err != nil { // io.EOF
		return
	}
	// number of arguments
	var argCount int
	if argCount, err = s.readLineInteger(); err != nil {
		return
	}

	cmd = s.reuseCommand(argCount)
	for i := 0; i < argCount; i++ {
		// Read ( $<number of bytes of argument 1> CR LF )
		err = s.skipSpecificByte('$')
		if err != nil {
			return
		}

		var argSize int
		argSize, err = s.readLineInteger()
		if err != nil {
			return
		}

		// Read ( <argument data> CR LF )
		var b []byte
		b, err = s.BlockReadBytes(argSize)
		if err != nil {
			return
		}
		cmd.Args = append(cmd.Args, b)

		err = s.skipSpecificBytes([]byte{CR, LF})
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

// 验证并跳过指定的字节，用于开始符和结束符的判断
func (s *Session) skipSpecificByte(c byte) (err error) {
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

func (s *Session) skipSpecificBytes(bs []byte) (err error) {
	for _, c := range bs {
		err = s.skipSpecificByte(c)
		if err != nil {
			break
		}
	}
	return
}

// 读取到Redis通用换行符为止
func (s *Session) readBytesToCRLF() (bs []byte, err error) {
	bs, err = s.lightReadBytes(CR)
	if err != nil {
		return
	}

	var c byte
	if c, err = s.rw.ReadByte(); err != nil {
		return
	} else if c != LF {
		err = errors.New(fmt.Sprintf("Illegal LF / %d", c))
	}

	return
}

// 读取字符串，遇到CRLF换行为止
func (s *Session) readLineString() (str string, err error) {
	var line []byte
	line, err = s.readBytesToCRLF()
	if err != nil {
		return
	}
	str = string(line)
	return
}

func (s *Session) ReadLineString() (str string, err error) {
	return s.readLineString()
}

// 读取整形，遇到CRLF换行为止
func (s *Session) readLineInteger() (i int, err error) {
	var line string
	line, err = s.readLineString()
	if err != nil {
		return
	}
	i, err = strconv.Atoi(line)
	return
}

func (s *Session) ReadLineInteger() (i int, err error) {
	return s.readLineInteger()
}

// Close conn
func (s *Session) Close() error {
	return s.conn.Close()
}

func (s *Session) String() string {
	return fmt.Sprintf("<Session:%s>", s.conn.RemoteAddr().String())
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

// 阻塞读取
func (s *Session) BlockReadBytes(bufsize int) (bs []byte, err error) {
	bs = make([]byte, bufsize)
	var n int
	n, err = s.rw.Read(bs)
	if err == io.EOF {
		return
	}
	// 如果网络较慢，会出现一次读不完，剩下的逐个读取
	if n < bufsize {
		//fmt.Printf("%d < bufsize %d\n", n, bufsize)
		var c byte
		for j := n; j < bufsize; j++ {
			c, err = s.rw.ReadByte()
			if err != nil {
				return
			}
			bs[j] = c
		}
	}
	return
}

// 简化的ReadBytes(delim)方法
// reader.ReadBytes(delim)创建对象过多，使用下面方法让GoRedis多处理2k/s
func (s *Session) lightReadBytes(delim byte) (line []byte, err error) {
	var c byte
	// cap=4，是因为大部分场景下，redis里的数据长度不大于9999
	line = make([]byte, 0, 4)
	for {
		c, err = s.rw.ReadByte()
		if err != nil {
			return
		}
		// 遇到结束符
		if c == delim {
			break
		}
		line = append(line, c)
	}
	return
}

func (s *Session) ReadRDB() (err error) {
	// Read ( $<number of bytes of RDB> CR LF )
	err = s.skipSpecificByte('$')
	if err != nil {
		return
	}

	var rdbSize int
	rdbSize, err = s.readLineInteger()
	if err != nil {
		return
	}

	for i := 0; i < rdbSize; i++ {
		_, err = s.rw.ReadByte()
		if err != nil {
			return
		}
	}
	return
}

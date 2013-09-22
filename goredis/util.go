// Copyright (c) 2013, Latermoon <lptmoon@gmail.com>
// All rights reserved.
// http://redis.io/topics/protocol
package goredis

import (
	"bufio"
	"errors"
	"io"
	"strconv"
)

// ==============================
// 各种工具方法
// ==============================

// 简化的ReadBytes(delim)方法
// reader.ReadBytes(delim)创建对象过多，使用下面方法让GoRedis多处理2k/s
func lightReadBytes(reader *bufio.Reader, delim byte) (line []byte, err error) {
	err = nil
	var c byte
	// cap=4，是因为大部分场景下，redis里的数据长度不大于9999
	line = make([]byte, 0, 4)
	for {
		c, err = reader.ReadByte()
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

func readInteger(reader *bufio.Reader, delim byte) (i int, err error) {
	var line []byte
	line, err = lightReadBytes(reader, delim)
	if err == nil {
		i, err = strconv.Atoi(string(line))
	}
	return
}

func ReadInteger(reader *bufio.Reader, delim byte) (i int, err error) {
	return readInteger(reader, delim)
}

// 阻塞读取
func blockReadBytes(reader *bufio.Reader, bufsize int) (bs []byte, err error) {
	bs = make([]byte, bufsize)
	var n int
	n, err = reader.Read(bs)
	if err == io.EOF {
		return
	}
	// 如果网络较慢，会出现一次读不完，剩下的逐个读取
	if n < bufsize {
		//fmt.Printf("%d < bufsize %d\n", n, bufsize)
		var c byte
		for j := n; j < bufsize; j++ {
			c, err = reader.ReadByte()
			if err != nil {
				return
			}
			bs[j] = c
		}
	}
	return
}

// 临时方法: 跳过指定长度
func SkipReaderCursor(reader *bufio.Reader, bufsize int) (err error) {
	for i := 0; i < bufsize; i++ {
		_, err = reader.ReadByte()
	}
	return
}

/*
In a Status Reply the first byte of the reply is "+"
In an Error Reply the first byte of the reply is "-"
In an Integer Reply the first byte of the reply is ":"
In a Bulk Reply the first byte of the reply is "$"
In a Multi Bulk Reply the first byte of the reply s "*"
*/
func ReadReply(reader *bufio.Reader) (reply *Reply, err error) {
	reply = &Reply{}
	var c byte
	if c, err = reader.ReadByte(); err != nil {
		return
	}

	switch c {
	case '+':
		reply.Type = ReplyTypeStatus
		reply.Value, err = lightReadBytes(reader, CR)
		_, err = reader.ReadBytes(LF) // CRLF
	case '-':
		reply.Type = ReplyTypeError
		reply.Value, err = lightReadBytes(reader, CR)
		_, err = reader.ReadBytes(LF) // CRLF
	case ':':
		reply.Type = ReplyTypeInteger
		reply.Value, err = readInteger(reader, CR)
		if err == nil {
			_, err = reader.ReadBytes(LF) // CRLF
		}
	case '$':
		reply.Type = ReplyTypeBulk
		var bufsize int
		bufsize, err = readInteger(reader, CR)
		if err != nil {
			break
		}
		_, err = reader.ReadBytes(LF) // CRLF
		reply.Value, err = blockReadBytes(reader, bufsize)
		if err != nil {
			break
		}
		_, err = reader.ReadBytes(LF) // CRLF
	case '*':
		reply.Type = ReplyTypeMultiBulks
		var argCount int
		argCount, err = readInteger(reader, CR)
		if err != nil {
			break
		}
		_, err = reader.ReadBytes(LF) // CRLF
		var c byte
		args := make([][]byte, argCount)
		for i := 0; i < argCount; i++ {
			c, err = reader.ReadByte()
			if err != nil || c != '$' {
				return
			}
			var argSize int
			argSize, err = readInteger(reader, CR)
			if err != nil {
				return
			}
			_, err = reader.ReadBytes(LF) // CRLF
			args[i], err = blockReadBytes(reader, argSize)
			_, err = reader.ReadBytes(LF) // CRLF
		}
		reply.Value = args
		_, err = reader.ReadBytes(LF) // CRLF
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
func ReadCommand(reader *bufio.Reader) (cmd *Command, err error) {
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
		var n int
		n, err = reader.Read(cmd.Args[i])
		if err == io.EOF {
			return
		}
		// 如果网络较慢，会出现一次读不完，剩下的逐个读取
		if n < argSize {
			//fmt.Printf("%d < argSize %d\n", n, argSize)
			var c byte
			for j := n; j < argSize; j++ {
				c, err = reader.ReadByte()
				if err != nil {
					return
				}
				cmd.Args[i][j] = c
			}
		}

		if c, err = reader.ReadByte(); err != nil {
			return
		} else if c != CR {
			err = errors.New("Illegal CR ..." + strconv.Itoa(argSize) + " " + string(c))
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

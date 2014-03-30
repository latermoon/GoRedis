// Copyright 2013 Latermoon. All rights reserved.

package goredis

import (
	"bytes"
	"errors"
	"fmt"
	"strconv"
)

// Command
type Command struct {
	Args [][]byte
}

func NewCommand(args ...[]byte) (cmd *Command) {
	cmd = &Command{
		Args: args,
	}
	return
}

// Name returns cmd.Args[0]
func (cmd *Command) Name() string {
	return string(cmd.Args[0])
}

func (cmd *Command) StringAtIndex(i int) string {
	if i >= len(cmd.Args) {
		return ""
	}
	return string(cmd.Args[i])
}

func (cmd *Command) ArgAtIndex(i int) (arg []byte, err error) {
	if i >= len(cmd.Args) {
		err = errors.New(fmt.Sprintf("out of range %d/%d", i, len(cmd.Args)))
		return
	}
	arg = cmd.Args[i]
	return
}

func (cmd *Command) IntAtIndex(i int) (n int, err error) {
	var f float64
	if f, err = cmd.FloatAtIndex(i); err == nil {
		n = int(f)
	}
	return
}

func (cmd *Command) Int64AtIndex(i int) (n int64, err error) {
	var f float64
	if f, err = cmd.FloatAtIndex(i); err == nil {
		n = int64(f)
	}
	return
}

func (cmd *Command) FloatAtIndex(i int) (n float64, err error) {
	if i >= len(cmd.Args) {
		err = errors.New(fmt.Sprintf("out of range %d/%d", i, len(cmd.Args)))
		return
	}
	n, err = strconv.ParseFloat(string(cmd.Args[i]), 64)
	return
}

func (cmd *Command) Len() int {
	return len(cmd.Args)
}

// Redis协议的Command数据
/*
*<number of arguments> CR LF
$<number of bytes of argument 1> CR LF
<argument data> CR LF
...
$<number of bytes of argument N> CR LF
<argument data> CR LF
*/
func (cmd *Command) Bytes() []byte {
	buf := bytes.Buffer{}
	buf.WriteByte('*')
	argCount := len(cmd.Args)
	//<number of arguments>
	buf.WriteString(itoa(argCount))
	buf.WriteString(CRLF)
	for i := 0; i < argCount; i++ {
		buf.WriteByte('$')
		//<number of bytes of argument i>
		argSize := len(cmd.Args[i])
		buf.WriteString(itoa(argSize))
		buf.WriteString(CRLF)
		buf.Write(cmd.Args[i]) //<argument data>
		buf.WriteString(CRLF)
	}
	return buf.Bytes()
}

// func ParseCommand(buf *bytes.Buffer) (*Command, error) {
// 	cmd := &Command{}

// 	// Read ( *<number of arguments> CR LF )
// 	if c, err := buf.ReadByte(); c != '*' { // io.EOF
// 		return nil, err
// 	}
// 	// number of arguments
// 	line, err := buf.ReadBytes(LF)
// 	if err != nil {
// 		return nil, err
// 	}
// 	argCount, _ := strconv.Atoi(string(line[:len(line)-2]))
// 	cmd.Args = make([][]byte, argCount)
// 	for i := 0; i < argCount; i++ {
// 		// Read ( $<number of bytes of argument 1> CR LF )
// 		if c, err := buf.ReadByte(); c != '$' {
// 			return nil, err
// 		}

// 		line, err := buf.ReadBytes(LF)
// 		if err != nil {
// 			return nil, err
// 		}
// 		argSize, _ := strconv.Atoi(string(line[:len(line)-2]))
// 		// Read ( <argument data> CR LF )
// 		cmd.Args[i] = make([]byte, argSize)
// 		n, e2 := buf.Read(cmd.Args[i])
// 		if n != argSize {
// 			return nil, errors.New("argSize too short")
// 		}
// 		if e2 != nil {
// 			return nil, e2
// 		}

// 		_, err = buf.ReadBytes(LF)
// 		if err != nil {
// 			return nil, err
// 		}
// 	}

// 	return cmd, nil
// }

func (cmd *Command) String() string {
	buf := &bytes.Buffer{}
	for i, count := 0, cmd.Len(); i < count; i++ {
		if i > 0 {
			buf.WriteString(" ")
		}
		buf.Write(cmd.Args[i])
	}
	return buf.String()
}

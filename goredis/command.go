// Copyright 2013 Latermoon. All rights reserved.

package goredis

import (
	"bytes"
	"errors"
	"fmt"
	"strconv"
)

// Command
type Command interface {
	Args() [][]byte // command data
	Len() int       // len(Args())
	Bytes() []byte  // Redis协议
	String() string
	// Accessor
	ArgAtIndex(i int) (arg []byte, err error)
	IntAtIndex(i int) (n int, err error)
	Int64AtIndex(i int) (n int64, err error)
	FloatAtIndex(i int) (n float64, err error)
	StringAtIndex(i int) string
	// Attribte
	SetAttribute(name string, v interface{})
	GetAttribute(name string) (v interface{})
}

// 内部实现
type baseCommand struct {
	args  [][]byte
	attrs map[string]interface{}
}

func NewCommand(args ...[]byte) (cmd Command) {
	cmd = &baseCommand{
		args:  args,
		attrs: make(map[string]interface{}),
	}
	return
}

func (cmd *baseCommand) SetAttribute(name string, v interface{}) {
	cmd.attrs[name] = v
}

func (cmd *baseCommand) GetAttribute(name string) (v interface{}) {
	return cmd.attrs[name]
}

func (cmd *baseCommand) Args() [][]byte {
	return cmd.args
}

func (cmd *baseCommand) StringAtIndex(i int) string {
	if i >= cmd.Len() {
		return ""
	}
	return string(cmd.args[i])
}

func (cmd *baseCommand) ArgAtIndex(i int) (arg []byte, err error) {
	if i >= cmd.Len() {
		err = errors.New(fmt.Sprintf("out of range %d/%d", i, cmd.Len()))
		return
	}
	arg = cmd.args[i]
	return
}

func (cmd *baseCommand) IntAtIndex(i int) (n int, err error) {
	var f float64
	if f, err = cmd.FloatAtIndex(i); err == nil {
		n = int(f)
	}
	return
}

func (cmd *baseCommand) Int64AtIndex(i int) (n int64, err error) {
	var f float64
	if f, err = cmd.FloatAtIndex(i); err == nil {
		n = int64(f)
	}
	return
}

func (cmd *baseCommand) FloatAtIndex(i int) (n float64, err error) {
	if i >= cmd.Len() {
		err = errors.New(fmt.Sprintf("out of range %d/%d", i, cmd.Len()))
		return
	}
	n, err = strconv.ParseFloat(string(cmd.args[i]), 64)
	return
}

func (cmd *baseCommand) Len() int {
	return len(cmd.args)
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
func (cmd *baseCommand) Bytes() []byte {
	buf := bytes.Buffer{}
	buf.WriteByte('*')
	argCount := cmd.Len()
	//<number of arguments>
	buf.WriteString(itoa(argCount))
	buf.WriteString(CRLF)
	for i := 0; i < argCount; i++ {
		buf.WriteByte('$')
		//<number of bytes of argument i>
		argSize := len(cmd.args[i])
		buf.WriteString(itoa(argSize))
		buf.WriteString(CRLF)
		buf.Write(cmd.args[i]) //<argument data>
		buf.WriteString(CRLF)
	}
	return buf.Bytes()
}

func ParseCommand(buf *bytes.Buffer) (Command, error) {
	cmd := &baseCommand{}

	// Read ( *<number of arguments> CR LF )
	if c, err := buf.ReadByte(); c != '*' { // io.EOF
		return nil, err
	}
	// number of arguments
	line, err := buf.ReadBytes(LF)
	if err != nil {
		return nil, err
	}
	argCount, _ := strconv.Atoi(string(line[:len(line)-2]))
	cmd.args = make([][]byte, argCount)
	for i := 0; i < argCount; i++ {
		// Read ( $<number of bytes of argument 1> CR LF )
		if c, err := buf.ReadByte(); c != '$' {
			return nil, err
		}

		line, err := buf.ReadBytes(LF)
		if err != nil {
			return nil, err
		}
		argSize, _ := strconv.Atoi(string(line[:len(line)-2]))
		// Read ( <argument data> CR LF )
		cmd.args[i] = make([]byte, argSize)
		n, e2 := buf.Read(cmd.args[i])
		if n != argSize {
			return nil, errors.New("argSize too short")
		}
		if e2 != nil {
			return nil, e2
		}

		_, err = buf.ReadBytes(LF)
		if err != nil {
			return nil, err
		}
	}

	return cmd, nil
}

func (cmd *baseCommand) String() string {
	buf := &bytes.Buffer{}
	for i, count := 0, cmd.Len(); i < count; i++ {
		if i > 0 {
			buf.WriteString(" ")
		}
		buf.Write(cmd.args[i])
	}
	return buf.String()
}

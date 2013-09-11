// Copyright (c) 2013, Latermoon <lptmoon@gmail.com>
// All rights reserved.
//
// 客户端指令
// @author latermoon
// @since 2013-08-27
package goredis

import (
	"bytes"
	"strconv"
)

// ==============================
// 代表一条客户端指令
// SET name Latermoon
// cmd.StringAtIndex(0) == cmd.Name() == "SET"
// cmd.StringAtIndex(1) == "name"
// cmd.StringAtIndex(2) == "Latermoon"
// ==============================
type Command struct {
	Args    [][]byte
	session *Session // 需要重构
}

func (cmd *Command) Session() *Session {
	return cmd.session
}

// 指令名称
// cmd.StringAtIndex(0) == cmd.Name() == "SET"
func (cmd *Command) Name() string {
	return string(cmd.Args[0])
}

// 参数按字符串返回
func (cmd *Command) StringAtIndex(i int) string {
	return string(cmd.Args[i])
}

// 返回全部参数的字符串形式
func (cmd *Command) StringArgs() (strs []string) {
	strs = make([]string, len(cmd.Args))
	for i, b := range cmd.Args {
		strs[i] = string(b)
	}
	return
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
	buf.WriteString("*")
	argCount := len(cmd.Args)
	buf.WriteString(strconv.Itoa(argCount)) //<number of arguments>
	buf.WriteString(CRLF)
	for i := 0; i < argCount; i++ {
		buf.WriteString("$")
		buf.WriteString(strconv.Itoa(len(cmd.Args[i]))) //<number of bytes of argument i>
		buf.WriteString(CRLF)
		buf.Write(cmd.Args[i]) //<argument data>
		buf.WriteString(CRLF)
	}
	return buf.Bytes()
}

func (cmd *Command) String() string {
	buf := bytes.Buffer{}
	for _, arg := range cmd.Args {
		buf.Write(arg)
		buf.WriteString(" ")
	}
	return buf.String()
}

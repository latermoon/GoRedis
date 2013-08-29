// 客户端指令Request
// @author Latermoon
// @since 2013-08-27
package goredis

import (
	"bytes"
	"fmt"
)

// ==============================
// 代表一条客户端指令
// 对于 SET name Latermoon
// cmd.StringAtIndex(0) == cmd.Name() == "SET"
// cmd.StringAtIndex(1) == "name"
// cmd.StringAtIndex(2) == "Latermoon"
// ==============================
type Command struct {
	Args [][]byte
}

// 指令的文本
// 客户端传入的数据，第一个参数就是指令，第二个参数开始是指令数据
func (cmd *Command) Name() string {
	return string(cmd.Args[0])
}

func (cmd *Command) ArgCount() int {
	return len(cmd.Args)
}

// 获取指定索引的参数
func (cmd *Command) ArgAtIndex(index int) (data []byte) {
	if index >= len(cmd.Args) {
		panic(fmt.Sprintf("Index Out of Bounds : %d/%d", index, len(cmd.Args)))
		return
	}
	data = cmd.Args[index]
	return
}

// 参数按字符串返回
func (cmd *Command) StringAtIndex(index int) (value string) {
	data := cmd.ArgAtIndex(index)
	value = string(data)
	return
}

func (cmd *Command) String() string {
	buf := bytes.Buffer{}
	for _, arg := range cmd.Args {
		buf.Write(arg)
		buf.WriteString(" ")
	}
	return buf.String()
}

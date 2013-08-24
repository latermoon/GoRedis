package goredis

import (
	"bytes"
	"errors"
	"fmt"
)

// ==============================
// 代表一条客户端指令
// ==============================
type Command struct {
	name interface{}
	Args [][]byte
}

// 指令的文本
// 客户端传入的数据，第一个参数就是指令，第二个参数开始是指令数据
func (cmd *Command) Name() string {
	if cmd.name == nil {
		cmd.name = string(cmd.Args[0])
	}
	return cmd.name.(string)
}

// 获取指定索引的参数
// 比如 SET name latermoon, cmd.ArgAtIndex(0) = SET, cmd.ArgAtIndex(1) = name, ...
func (cmd *Command) ArgAtIndex(index int) (data []byte, err error) {
	if index >= len(cmd.Args) {
		err = errors.New(fmt.Sprintf("Index Out of Bounds : %d/%d", index, len(cmd.Args)))
		return
	}
	err = nil
	data = cmd.Args[index]
	return
}

// 参数按字符串返回
func (cmd *Command) StringAtIndex(index int) (value string, err error) {
	data, err := cmd.ArgAtIndex(index)
	if err != nil {
		return
	}
	value = string(data)
	return
}

func (cmd *Command) String() string {
	buf := bytes.Buffer{}
	argCount := len(cmd.Args)
	for i := 0; i < argCount; i++ {
		buf.Write(cmd.Args[i])
		buf.WriteString(" ")
	}
	return buf.String()
}

package goredis_server

import (
	. "GoRedis/goredis"
	"errors"
	"strings"
)

// 验证指令是否合法，包括：
// 传入的参数数量，key里是否包含非法字符
var (
	BadCommandError    = errors.New("bad command")
	WrongArgumentCount = errors.New("wrong argument count")
	WrongCommandKey    = errors.New("wrong command key")
)

var (
	cmdrules = make(map[string][]interface{})
)

// RuleIndex
const (
	RI_MinCount  = iota
	RI_MaxCount  // -1 for undefined
	RI_OddOrEven // 0 for Odd, 1 for Even, -1 for undefined
)

func init() {
	cmdrules = map[string][]interface{}{
		"GET":    []interface{}{2, 2},
		"SET":    []interface{}{3, -1},
		"MGET":   []interface{}{2, -1},
		"MSET":   []interface{}{3, -1},
		"INCR":   []interface{}{2, 2},
		"DECR":   []interface{}{2, 2},
		"INCRBY": []interface{}{3, 3},
		"DECRBY": []interface{}{3, 3},
	}
}

func verifyCommand(cmd *Command) error {
	if cmd == nil || cmd.Len() == 0 {
		return BadCommandError
	}

	name := strings.ToUpper(cmd.Name())
	rule, exist := cmdrules[name]
	if !exist {
		return nil
	}

	for i, count := 0, len(rule); i < count; i++ {
		switch i {
		case RI_MinCount:
			if cmd.Len() < rule[i].(int) {
				return WrongArgumentCount
			}
		case RI_MaxCount:
			if cmd.Len() > rule[i].(int) {
				return WrongArgumentCount
			}
		}
	}

	if cmd.Len() > 1 {
		key := cmd.StringAtIndex(1)
		if strings.ContainsAny(key, "#[] ") {
			return WrongCommandKey
		}
	}
	return nil
}

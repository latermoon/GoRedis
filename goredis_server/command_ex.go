package goredis_server

import (
	. "GoRedis/goredis"
	"strings"
	"time"
)

// 扩展Command的业务功能
type CommandEx struct {
	*Command
	session   *Session
	upperName string
	execTime  time.Duration
}

func NewCommandEx(session *Session, cmd *Command) (c *CommandEx) {
	c = &CommandEx{
		session: session,
		Command: cmd,
	}
	return
}

func (c *CommandEx) Session() *Session {
	return c.session
}

//  验证指令是有合法
func (c *CommandEx) Verify() error {
	return verifyCommand(c.Command)
}

// 是否需要同步
func (c *CommandEx) NeedSync() bool {
	return needSync(c.UpperName())
}

// 记录执行耗时
func (c *CommandEx) SetExecTime(d time.Duration) {
	c.execTime = d
}

func (c *CommandEx) ExecTime() time.Duration {
	return c.execTime
}

// 大写名字
func (c *CommandEx) UpperName() string {
	if len(c.upperName) == 0 {
		c.upperName = strings.ToUpper(c.Command.Name())
	}
	return c.upperName
}

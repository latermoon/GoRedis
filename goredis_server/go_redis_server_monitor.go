package goredis_server

import (
	. "GoRedis/goredis"
	"strings"
)

// redis-cli对monitor指令进行特殊处理，只要monitor不断输出StatusReply，可以实现不间断的流输出
// 适用于海量数据的扫描输出，比如iterator扫描整个数据库
func (server *GoRedisServer) OnMONITOR(session *Session, cmd *Command) (reply *Reply) {
	// 特殊使用，monitor输出全部key
	if len(cmd.Args) > 1 {
		switch strings.ToLower(cmd.StringAtIndex(1)) {
		case "keys":
			server.monitorKeys(session, cmd)
			return
		case "sync":
			cmd = NewCommand([]byte("SYNC"), []byte(""))
			server.OnSYNC(session, cmd)
			return
		}
	}

	session.WriteReply(StatusReply("OK"))
	client := NewMonClient(session)
	server.monmgr.Add(client)
	return
}

// echo 'monitor keys' | redis-cli -p 1602
func (server *GoRedisServer) monitorKeys(session *Session, cmd *Command) {
	prefix, _ := cmd.ArgAtIndex(2)
	server.levelRedis.Keys(prefix, func(i int, key, keytype []byte, quit *bool) {
		err := session.WriteReply(StatusReply(string(key)))
		if err != nil {
			*quit = true
		}
	})
	session.Close()
}

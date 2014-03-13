package goredis_server

import (
	. "GoRedis/goredis"
	"GoRedis/libs/stdlog"
	"strings"
	"time"
)

// redis-cli对monitor指令进行特殊处理，只要monitor不断输出StatusReply，可以实现不间断的流输出
// 适用于海量数据的扫描输出，比如iterator扫描整个数据库
func (server *GoRedisServer) OnMONITOR(session *Session, cmd *Command) (reply *Reply) {
	// 特殊使用，monitor输出全部key
	if cmd.Len() > 1 {
		switch strings.ToUpper(cmd.StringAtIndex(1)) {
		case "KEYS":
			server.monitorKeys(session, cmd)
		default:
			reply = ErrorReply("bad monitor command")
			go func() {
				time.Sleep(time.Millisecond * 100)
				session.Close()
			}()
		}
		return
	}

	session.WriteReply(StatusReply("OK"))
	client := NewMonClient(session)
	remoteHost := session.RemoteAddr().String()

	go func() {
		stdlog.Printf("[%s] monitor start\n", remoteHost)
		// sleep一下，避免启动瞬间输出 +1394530022.495448 [0 127.0.0.1:51980] "monitor"
		time.Sleep(time.Millisecond * 10)
		server.monmgr.Put(remoteHost, client)
		client.Start()
		server.monmgr.Remove(remoteHost)
		stdlog.Printf("[%s] monitor exit\n", remoteHost)
	}()

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

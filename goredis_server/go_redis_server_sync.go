package goredis_server

import (
	. "GoRedis/goredis"
	"GoRedis/libs/levelredis"
	"GoRedis/libs/stdlog"
	"bytes"
	"strings"
)

// 向从库发送数据
// SYNC uid 70ecc21580
// 对应 go_redis_server_slaveof.go
func (server *GoRedisServer) OnSYNC(session *Session, cmd *Command) (reply *Reply) {
	// 客户端标识 SYNC uid 70ecc21580
	uid := ""
	args := cmd.StringArgs()
	if len(args) >= 3 && strings.ToLower(args[1]) == "uid" {
		uid = args[2]
	}
	stdlog.Printf("[%s] new slave uid %s\n", session.RemoteAddr(), uid)

	sc, err := NewSyncClient(session, server.directory)
	if err != nil {
		stdlog.Printf("[%s] new slave error %s", session.RemoteAddr(), err)
		return
	}
	server.syncmgr.Add(sc)
	stdlog.Printf("[%s] start send snapshot\n", session.RemoteAddr())
	go server.sendSnapshot(sc)

	return // SYNC不需要Reply
}

func (server *GoRedisServer) sendSnapshot(sc *SyncClient) {
	cmdName := []byte("RAW_SET_NOREPLY")
	server.levelRedis.SnapshotEnumerate([]byte{}, []byte{levelredis.MAXBYTE}, func(i int, key, value []byte, quit *bool) {
		if bytes.HasPrefix(key, []byte(goredisPrefix)) {
			return
		}
		cmd := NewCommand(cmdName, key, value)
		err := sc.Send(cmd)
		if err != nil {
			stdlog.Printf("[%s] send snapshot error %s\n", sc.session.RemoteAddr(), cmd)
			*quit = true
		}
		// stdlog.Printf("snapshot: %s,%s\n", string(key), string(value))
	})
	stdlog.Printf("[%s] send snapshot finish\n", sc.session.RemoteAddr())
	if sc.Available() {
		sc.StartSync()
	}
}

package goredis_server

import (
	. "../goredis"
	. "./storage"
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

	slave := server.findSlaveById(uid)
	if slave == nil {
		server.stdlog.Info("[%s] new slave %s", uid, session.RemoteAddr())
		snapshot, err := server.datasource.(*LevelDBDataSource).DB().GetSnapshot()
		if err != nil {
			return ErrorReply(err)
		}
		slave = NewSlaveSession(server, session, uid)
		server.slavelist.PushBack(slave)
		go slave.SendSnapshot(snapshot)
	} else {
		server.stdlog.Info("[%s] slave already exists", uid)
		slave.SetSession(session)
		go slave.ContinueSync()
	}
	// SYNC不需要Reply
	reply = nil
	return
}

func (server *GoRedisServer) findSlaveById(uid string) (slave *SlaveSession) {
	if len(uid) == 0 {
		return
	}
	for e := server.slavelist.Front(); e != nil; e = e.Next() {
		if e.Value.(*SlaveSession).UID() == uid {
			slave = e.Value.(*SlaveSession)
			return
		}
	}
	return
}

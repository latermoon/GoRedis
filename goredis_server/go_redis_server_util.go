package goredis_server

import (
	. "../goredis"
	"os"
	"runtime"
	"runtime/pprof"
	"strconv"
)

func (server *GoRedisServer) formatFloat(f float64) string {
	return strconv.FormatFloat(f, 'g', 12, 64)
}

func (server *GoRedisServer) OnGC(cmd *Command) (reply *Reply) {
	runtime.GC()
	reply = StatusReply("OK")
	return
}

func (server *GoRedisServer) OnPPROF(cmd *Command) (reply *Reply) {
	action := cmd.StringAtIndex(1)
	switch action {
	case "mem":
		f, err := os.Create(server.directory + "mem.prof")
		if err != nil {
			return ErrorReply(err)
		}
		pprof.WriteHeapProfile(f)
		f.Close()
		reply = StatusReply("OK")
	case "stop2":
		reply = StatusReply("OK")
	default:
		reply = ErrorReply("pprof [start/stop]")
	}
	return
}

func (server *GoRedisServer) OnPPROF_START(cmd *Command) (reply *Reply) {

	reply = StatusReply("OK")
	return
}

// // 存放id
// func (server *GoRedisServer) slaveIdMap() (m map[string]interface{}) {
// 	m = server.config.GetMap("slaveids")
// 	if m == nil {
// 		m = make(map[string]interface{})
// 	}
// 	return
// }

// // 发送snapshot完成后的回调
// func (server *GoRedisServer) snapshotSentCallback(session *SlaveSession) {
// 	m := server.slaveIdMap()
// 	if session.AofEnabled() {
// 		m[session.UID()] = ""
// 	}
// 	server.config.SetMap("slaveids", m)
// }

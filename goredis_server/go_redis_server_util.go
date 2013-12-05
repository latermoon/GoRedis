package goredis_server

import (
	"strconv"
)

func (server *GoRedisServer) formatFloat(f float64) string {
	return strconv.FormatFloat(f, 'g', 12, 64)
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

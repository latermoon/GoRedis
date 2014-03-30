package goredis_proxy

import (
	"GoRedis/libs/stdlog"
)

// 初始化入口
func (server *GoRedisProxy) Init() (err error) {
	e1 := server.resetMaster(server.options.MasterHost)
	e2 := server.resetSlave(server.options.SlaveHost)
	if e1 != nil {
		stdlog.Println(e1)
	}
	if e2 != nil {
		stdlog.Println(e2)
	}
	return
}

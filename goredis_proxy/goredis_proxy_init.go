package goredis_proxy

// 初始化入口
func (server *GoRedisProxy) Init() (err error) {
	err = server.resetMaster(server.options.MasterHost)
	if err != nil {
		return
	}
	err = server.resetSlave(server.options.SlaveHost)
	if err != nil {
		return
	}
	return
}

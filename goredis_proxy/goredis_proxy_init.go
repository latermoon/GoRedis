package goredis_proxy

// 初始化入口
func (server *GoRedisProxy) Init() (err error) {
	server.master, err = NewRemoteSession(server.options.MasterHost)
	if err != nil {
		return
	}
	server.slave, err = NewRemoteSession(server.options.SlaveHost)
	if err != nil {
		return
	}
	return
}

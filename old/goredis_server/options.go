package goredis_server

// 运行配置
type Options struct {
	host        string
	port        int
	dbpath      string
	logpath     string
	slaveofHost string
	slaveofPort int
}

func NewOptions() (o *Options) {
	o = &Options{}
	return
}

func (o *Options) SetHost(host string) {
	o.host = host
}

func (o *Options) Host() string {
	return o.host
}

func (o *Options) SetPort(port int) {
	o.port = port
}

func (o *Options) Port() int {
	return o.port
}

func (o *Options) SetDBPath(dbpath string) {
	o.dbpath = dbpath
}

func (o *Options) DBPath() string {
	return o.dbpath
}

func (o *Options) SetLogPath(logpath string) {
	o.logpath = logpath
}

func (o *Options) LogPath() string {
	return o.logpath
}

func (o *Options) SetSlaveOf(host string, port int) {
	o.slaveofHost, o.slaveofPort = host, port
}

func (o *Options) SlaveOf() (host string, port int) {
	return o.slaveofHost, o.slaveofPort
}

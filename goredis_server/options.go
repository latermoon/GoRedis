package goredis_server

// 运行配置
type Options struct {
	bind        string
	slaveofHost string
	slaveofPort int
	directory   string
	logdir      string
}

func NewOptions() (o *Options) {
	o = &Options{}
	return
}

func (o *Options) SetBind(host string) {
	o.bind = host
}

func (o *Options) Bind() string {
	return o.bind
}

func (o *Options) SetSlaveOf(host string, port int) {
	o.slaveofHost, o.slaveofPort = host, port
}

func (o *Options) SlaveOf() (host string, port int) {
	return o.slaveofHost, o.slaveofPort
}

func (o *Options) SetDirectory(path string) {
	o.directory = path
}

func (o *Options) Directory() string {
	return o.directory
}

func (o *Options) LogDir() string {
	return o.logdir
}

func (o *Options) SetLogDir(logdir string) {
	o.logdir = logdir
}

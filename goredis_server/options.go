package goredis_server

type Options struct {
	bind      string
	slaveof   string
	directory string
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

func (o *Options) SetSlaveOf(host string) {
	o.slaveof = host
}

func (o *Options) SlaveOf() string {
	return o.slaveof
}

func (o *Options) SetDirectory(path string) {
	o.directory = path
}

func (o *Options) Directory() string {
	return o.directory
}

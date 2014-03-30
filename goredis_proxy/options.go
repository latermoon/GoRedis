package goredis_proxy

import (
	"fmt"
)

type Options struct {
	MasterHost string
	SlaveHost  string
	Host       string
	Port       int
	Mode       string
	PoolSize   int
}

func NewOptions() (o *Options) {
	o = &Options{
		Port:     1602,
		PoolSize: 100,
	}
	return
}

func (o *Options) Addr() string {
	return fmt.Sprintf("%s:%d", o.Host, o.Port)
}

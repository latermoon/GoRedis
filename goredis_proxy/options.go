package goredis_proxy

import (
	"fmt"
)

type Options struct {
	MasterHost string
	SlaveHost  string
	Host       string
	Port       int
}

func (o *Options) Addr() string {
	return fmt.Sprintf("%s:%d", o.Host, o.Port)
}

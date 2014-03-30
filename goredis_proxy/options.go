package goredis_proxy

import (
	"fmt"
	"strings"
)

// Mode=r, 从库提供读，写操作返回错误
// Mode=rr, 主从均提供读，写操作返回错误
// Mode=rw, 主库提供写，从库提供读
// Mode=rrw，主库提供读写，从库提供读
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
		Host:     "",
		Port:     1602,
		Mode:     "rw",
		PoolSize: 100,
	}
	return
}

func (o *Options) Addr() string {
	return fmt.Sprintf("%s:%d", o.Host, o.Port)
}

// 是否允许写主库
func (o *Options) CanWrite() bool {
	return strings.Contains(o.Mode, "w")
}

func (o *Options) CanReadMaster() bool {
	return strings.Contains(o.Mode, "rr")
}

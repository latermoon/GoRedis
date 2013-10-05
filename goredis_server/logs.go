package goredis_server

import (
	"./libs/log4go"
)

// 全局变量
var (
	stdlog log4go.Logger
)

func init() {
	stdlog = log4go.NewDefaultLogger(log4go.DEBUG)
}

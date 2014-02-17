package goredis_server

import (
	"GoRedis/libs/stdlog"
	"os"
	"os/signal"
	"syscall"
)

// 处理退出事件
func (server *GoRedisServer) initSignalNotify() {
	server.sigs = make(chan os.Signal, 1)
	signal.Notify(server.sigs, syscall.SIGTERM)
	go func() {
		sig := <-server.sigs
		server.levelRedis.Close()
		stdlog.Println("signal:", sig)
		os.Exit(0)
	}()
}

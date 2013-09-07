package main

import (
	"../goredis"
	"fmt"
	"runtime"
)

// 简单的Redis服务器处理函数
type SimpleServerHandler struct {
	goredis.CommandHandler
}

func (s *SimpleServerHandler) On(name string, cmd *goredis.Command) (reply *goredis.Reply) {
	reply = goredis.ErrorReply("Not Supported:" + name)
	return
}

func (s *SimpleServerHandler) OnGET(cmd *goredis.Command) (reply *goredis.Reply) {
	key := cmd.StringAtIndex(1)
	reply = goredis.BulkReply(key)
	return
}

func (s *SimpleServerHandler) OnSET(cmd *goredis.Command) (reply *goredis.Reply) {
	reply = goredis.StatusReply("OK")
	return
}

func main() {
	runtime.GOMAXPROCS(2)
	fmt.Println("SimpleServer start, listen 1603 ...")
	server := goredis.NewRedisServer(&SimpleServerHandler{})
	server.Listen(":1603")
}

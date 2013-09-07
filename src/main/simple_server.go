package main

import (
	. "../goredis"
	"fmt"
	"runtime"
)

// 简单的Redis服务器处理函数
type SimpleServerHandler struct {
	CommandHandler
}

func (s *SimpleServerHandler) On(name string, cmd *Command) (reply *Reply) {
	reply = ErrorReply("Not Supported: " + cmd.String())
	return
}

func (s *SimpleServerHandler) OnGET(cmd *Command) (reply *Reply) {
	key := cmd.StringAtIndex(1)
	reply = BulkReply("value of " + key)
	return
}

func (s *SimpleServerHandler) OnSET(cmd *Command) (reply *Reply) {
	key := cmd.StringAtIndex(1)
	val := cmd.StringAtIndex(2)
	reply = StatusReply("OK, " + key + "=" + val)
	return
}

func (s *SimpleServerHandler) OnINFO(cmd *Command) (reply *Reply) {
	lines := "Powerby GoRedis" + "\n"
	lines += "SimpleRedisServer" + "\n"
	lines += "Support GET/SET/INFO" + "\n"
	reply = BulkReply(lines)
	return
}

func main() {
	runtime.GOMAXPROCS(2)
	fmt.Println("SimpleServer start, listen 1603 ...")
	server := NewRedisServer(&SimpleServerHandler{})
	server.Listen(":1603")
}

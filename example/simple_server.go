package main

import (
	"fmt"
	. "github.com/latermoon/GoRedis/src/goredis"
	"runtime"
)

// ==============================
// 简单的Redis服务器处理类
// ==============================
type SimpleServerHandler struct {
	CommandHandler
	kvCache map[string]interface{} // KeyValue
	kvLock  chan int               // Set操作的写锁
}

func NewSimpleServerHandler() (handler *SimpleServerHandler) {
	handler = &SimpleServerHandler{}
	handler.kvCache = make(map[string]interface{})
	handler.kvLock = make(chan int, 1)
	return
}

func (s *SimpleServerHandler) On(name string, cmd *Command) (reply *Reply) {
	reply = ErrorReply("Not Supported: " + cmd.String())
	return
}

func (s *SimpleServerHandler) OnGET(cmd *Command) (reply *Reply) {
	key := cmd.StringAtIndex(1)
	value := s.kvCache[key]
	reply = BulkReply(value)
	return
}

func (s *SimpleServerHandler) OnSET(cmd *Command) (reply *Reply) {
	key := cmd.StringAtIndex(1)
	value := cmd.StringAtIndex(2)
	s.kvLock <- 0
	s.kvCache[key] = value
	<-s.kvLock
	reply = StatusReply("OK")
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
	server := NewRedisServer(NewSimpleServerHandler())
	server.Listen(":1603")
}

package main

import (
	. "../goredis"
	"flag"
	"fmt"
	"runtime"
	"sync"
)

// ==============================
// 简单的Redis服务器处理类
// ==============================
type SimpleServerHandler struct {
	CommandHandler
	cache map[string]interface{} // KeyValue
	mu    sync.Mutex             // Set操作的写锁
}

func NewSimpleServerHandler() (handler *SimpleServerHandler) {
	handler = &SimpleServerHandler{}
	handler.cache = make(map[string]interface{})
	return
}

// 跟踪全部指令，没有reply，只用于日志等用途
func (s *SimpleServerHandler) On(session *Session, cmd *Command) {
	return
}

// 处理未知指令触发
func (s *SimpleServerHandler) OnUndefined(session *Session, cmd *Command) (reply *Reply) {
	return ErrorReply("Not Supported: " + cmd.String())
}

// 处理特定指令
func (s *SimpleServerHandler) OnGET(cmd *Command) (reply *Reply) {
	key := cmd.StringAtIndex(1)
	value := s.cache[key]
	reply = BulkReply(value)
	return
}

func (s *SimpleServerHandler) OnSET(cmd *Command) (reply *Reply) {
	s.mu.Lock()
	defer s.mu.Unlock()
	key := cmd.StringAtIndex(1)
	value := cmd.StringAtIndex(2)
	// set
	s.cache[key] = value
	reply = StatusReply("OK")
	return
}

func (s *SimpleServerHandler) OnINFO(cmd *Command) (reply *Reply) {
	lines := "Powerby GoRedis" + "\n"
	lines += "SimpleRedisServer" + "\n"
	lines += "Only Support GET/SET/INFO" + "\n"
	reply = BulkReply(lines)
	return
}

func main() {
	runtime.GOMAXPROCS(2)
	//flag
	portPtr := flag.Int("p", 1601, "Server port")
	flag.Parse()
	fmt.Printf("SimpleServer start, listen %d ...\r\n", *portPtr)

	server := NewRedisServer(NewSimpleServerHandler())
	server.Listen(fmt.Sprintf(":%d", *portPtr))
}

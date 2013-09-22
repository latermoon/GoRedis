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
	kvCache map[string]interface{} // KeyValue
	kvMutex *sync.Mutex            // Set操作的写锁
}

func NewSimpleServerHandler() (handler *SimpleServerHandler) {
	handler = &SimpleServerHandler{}
	handler.kvCache = make(map[string]interface{})
	handler.kvMutex = &sync.Mutex{}
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
	s.kvMutex.Lock()
	s.kvCache[key] = value
	s.kvMutex.Unlock()
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
	//flag
	portPtr := flag.Int("p", 1601, "Server port")
	flag.Parse()
	fmt.Printf("SimpleServer start, listen %d ...\r\n", *portPtr)

	cmd := NewCommand([]byte("SYNC"), []byte("7b0a1520"))
	fmt.Println(cmd.String())

	server := NewRedisServer(NewSimpleServerHandler())
	server.Listen(fmt.Sprintf(":%d", *portPtr))
}

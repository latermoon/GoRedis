// ==============================
// RedisServer实例
// 实现最原始的Handle来处理请求
// 安装方式：
// 配置$GOPATH后
// install: go get github.com/latermoon/GoRedis/goredis
// update: go get -u github.com/latermoon/GoRedis/goredis
// user: import "github.com/latermoon/GoRedis/goredis"
// ==============================
package main

import (
	"fmt"
	//"github.com/latermoon/GoRedis/goredis"
	"../goredis"
	//"runtime"
)

func main() {
	fmt.Println("GoRedis 0.1 by latermoon")
	//runtime.GOMAXPROCS(1)
	server := goredis.NewRedisServer()

	// KeyValue
	kvCache := make(map[string]interface{})
	// Set操作的写锁
	chanSet := make(chan int, 1)

	server.On("GET", func(cmd *goredis.Command) (reply *goredis.Reply) {
		key := cmd.StringAtIndex(1)
		value := kvCache[key]
		reply = goredis.BulkReply(value)
		return
	})

	server.On("SET", func(cmd *goredis.Command) (reply *goredis.Reply) {
		key := cmd.StringAtIndex(1)
		value := cmd.StringAtIndex(2)
		chanSet <- 0
		kvCache[key] = value
		<-chanSet
		reply = goredis.StatusReply("OK")
		return
	})

	server.On("PING", func(cmd *goredis.Command) (reply *goredis.Reply) {
		reply = goredis.StatusReply("PONG")
		return
	})

	server.On("INFO", func(cmd *goredis.Command) (reply *goredis.Reply) {
		reply = goredis.BulkReply("GoRedis 0.1 by latermoon\n")
		return
	})

	// 开始监听端口
	fmt.Println("Listen :8002")
	server.Listen(":8002")
}

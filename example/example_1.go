// ==============================
// BaseServer实例
// 实现最原始的Handle来处理请求
// ==============================
package main

import (
	"github.com/latermoon/GoRedis/goredis"
	//"container/list"
	"fmt"
)

func main() {
	fmt.Println("GoRedis 0.1 by latermoon")

	server, _ := goredis.NewRedisServer()

	// KeyValue
	kvCache := make(map[string]interface{})
	// Set操作的写锁
	chanSet := make(chan int, 1)

	server.On("GET", func(session *goredis.Session, cmd *goredis.Command) (err error) {
		err = nil
		key, _ := cmd.StringAtIndex(1)
		value := kvCache[key]
		session.ReplyBulk(value)
		return
	})

	server.On("SET", func(session *goredis.Session, cmd *goredis.Command) (err error) {
		key, _ := cmd.StringAtIndex(1)
		value, _ := cmd.StringAtIndex(2)
		chanSet <- 0
		kvCache[key] = value
		<-chanSet
		session.ReplyStatus("OK")
		return
	})

	server.On("INFO", func(session *goredis.Session, cmd *goredis.Command) (err error) {
		err = nil
		session.ReplyBulk("GoRedis 0.1 by latermoon\n")
		return
	})

	// 开始监听端口
	server.Listen(":8002")
}

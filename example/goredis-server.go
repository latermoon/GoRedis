// goredis-server启动函数
// @author latermoon

package main

import (
	"../goredis_server"
	"flag"
	"fmt"
	"runtime"
)

// go run goredis-server.go -h localhost -p 1602
func main() {
	runtime.GOMAXPROCS(4)
	hostPtr := flag.String("h", "localhost", "Server host")
	portPtr := flag.Int("p", 1602, "Server port")
	flag.Parse()

	host := fmt.Sprintf("%s:%d", *hostPtr, *portPtr)
	fmt.Println("GoRedis 0.1 by latermoon")
	fmt.Println("Listen", host)

	directory := fmt.Sprintf("/tmp/goredis_%d/", *portPtr)
	server := goredis_server.NewGoRedisServer(directory)
	server.Listen(host)
}

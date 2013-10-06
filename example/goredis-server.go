// goredis-server启动函数
// @author latermoon

package main

import (
	"../goredis_server"
	"flag"
	"fmt"
	"os"
	"runtime"
)

// go run goredis-server.go -h localhost -p 1602
func main() {
	runtime.GOMAXPROCS(4)
	fmt.Println("GoRedis 0.1 by latermoon")

	hostPtr := flag.String("h", "", "Server host")
	portPtr := flag.Int("p", 1602, "Server port")
	flag.Parse()

	host := fmt.Sprintf("%s:%d", *hostPtr, *portPtr)

	directory := fmt.Sprintf("/tmp/goredis_%d/", *portPtr)
	os.MkdirAll(directory, os.ModePerm)
	server := goredis_server.NewGoRedisServer(directory)
	server.Init()
	server.Listen(host)
}

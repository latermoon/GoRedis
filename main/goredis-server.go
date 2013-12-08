// goredis-server启动函数
// @author latermoon

package main

import (
	"../goredis_server"
	stdlog "../libs/stdlog"
	"flag"
	"fmt"
	"os"
	"runtime"
)

// go run goredis-server.go -h localhost -p 1602
// go run goredis-server.go -procs 8 -p 17600
func main() {
	version := flag.Bool("v", false, "print goredis-server version")
	hostPtr := flag.String("h", "", "server host")
	portPtr := flag.Int("p", 1602, "server port")
	procsPtr := flag.Int("procs", 8, "GOMAXPROCS")
	flag.Parse()

	if *version {
		stdlog.Println("goredis-server", goredis_server.VERSION)
		return
	}

	runtime.GOMAXPROCS(*procsPtr)

	// db parent directory
	dbhome := "/data"
	finfo, e1 := os.Stat(dbhome)
	if os.IsNotExist(e1) || !finfo.IsDir() {
		dbhome = "/tmp"
	}

	directory := fmt.Sprintf("%s/goredis_%d/", dbhome, *portPtr)
	os.MkdirAll(directory, os.ModePerm)

	host := fmt.Sprintf("%s:%d", *hostPtr, *portPtr)

	// start ...
	server := goredis_server.NewGoRedisServer(directory)
	server.Init()
	server.Listen(host)
}

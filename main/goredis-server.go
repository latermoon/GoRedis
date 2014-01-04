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
	"time"
)

func init() {
	// 全局日志前缀
	stdlog.SetPrefix(func() string {
		t := time.Now()
		return fmt.Sprintf("[%d-%02d-%02d %02d:%02d:%02d] ", t.Year(), t.Month(), t.Day(), t.Hour(), t.Minute(), t.Second())
	})
}

// go run goredis-server.go -h localhost -p 1602
// go run goredis-server.go -procs 8 -p 17600
func main() {
	version := flag.Bool("v", false, "print goredis-server version")
	hostPtr := flag.String("h", "", "server host")
	portPtr := flag.Int("p", 1602, "server port")
	procsPtr := flag.Int("procs", 8, "GOMAXPROCS")
	flag.Parse()

	if *version {
		fmt.Println("goredis-server", goredis_server.VERSION)
		return
	}

	runtime.GOMAXPROCS(*procsPtr)

	// 设置主路径
	dbhome := "/data"
	finfo, e1 := os.Stat(dbhome)
	if os.IsNotExist(e1) || !finfo.IsDir() {
		dbhome = "/tmp"
	}
	directory := fmt.Sprintf("%s/goredis_%d/", dbhome, *portPtr)
	os.MkdirAll(directory, os.ModePerm)

	// 设置全局输出路径
	out, err := os.OpenFile(directory+"stdout.log", os.O_WRONLY|os.O_CREATE|os.O_APPEND, os.ModePerm)
	if err != nil {
		panic(err)
	}
	// 同时输出到屏幕和文件
	stdlog.SetOutput(stdlog.NewMultiWriter(os.Stdout, out))

	host := fmt.Sprintf("%s:%d", *hostPtr, *portPtr)

	stdlog.Println("========================================")
	// start ...
	server := goredis_server.NewGoRedisServer(directory)
	server.Init()
	server.Listen(host)
}

// goredis-slaveof 主从同步代理
// 典型应用场景是异地机房redis主从同步，master和goredis-slaveof在同一个机房，保障稳定连接
// 然后goredis-slaveof把master的数据队列到本地，然后传输到异地机房slave实例，专线断开不会导致全量主从同步
// @author latermoon

package main

import (
	. "./tool/slaveof"
	"flag"
	"fmt"
	"os"
	"runtime"
)

// go run goredis-slaveof.go -src host:port -dst host:port
func main() {
	version := flag.Bool("v", false, "print goredis-slaveof version")
	srcPtr := flag.String("src", "", "source host")
	dstPtr := flag.String("dst", "", "dest host")
	procsPtr := flag.Int("procs", 8, "GOMAXPROCS")
	flag.Parse()

	if *version {
		fmt.Println("goredis-slaveof", "0.1.2")
		return
	}

	runtime.GOMAXPROCS(*procsPtr)

	// db parent directory
	dbhome := "/data"
	finfo, e1 := os.Stat(dbhome)
	if os.IsNotExist(e1) || !finfo.IsDir() {
		dbhome = "/tmp"
	}

	if len(*dstPtr) == 0 || len(*srcPtr) == 0 {
		panic("bad dsthost or srchost")
	}

	directory := fmt.Sprintf("%s/goredis_%s_slaveof_%s/", dbhome, *dstPtr, *srcPtr)
	fmt.Println("dbhome:", directory)

	slaveClient, err := NewSlaveOf(directory, *srcPtr, *dstPtr)
	if err != nil {
		panic(err)
	}
	slaveClient.Start()
}

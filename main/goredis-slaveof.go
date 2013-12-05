// goredis-slaveof从库中转站
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
		fmt.Println("goredis-slaveof", "0.1.1")
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

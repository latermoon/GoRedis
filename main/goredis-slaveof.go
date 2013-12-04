// goredis-slaveof从库中转站
// @author latermoon

package main

import (
	"../goredis_server"
	"flag"
	"fmt"
	"os"
	"runtime"
)

// go run goredis-slaveof.go -src host:port -dst host:port
func main() {
	version := flag.Bool("v", false, "print goredis-slaveof version")
	srcPtr := flag.String("src", "", "source host")
	dstPtr := flag.Int("dst", 1602, "dest host")
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

	directory := fmt.Sprintf("%s/goredis_%s_slaveof_%s/", dbhome, *dstPtr, *srcPtr)
	os.MkdirAll(directory, os.ModePerm)

}

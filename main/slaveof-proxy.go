package main

import (
	. "./tool/slaveof"
	"GoRedis/libs/stdlog"
	"flag"
	"fmt"
	"runtime"
	"time"
)

// ./slaveof-proxy -src localhost:6379 -dest remote:6379 -pullrate 400
func main() {
	runtime.GOMAXPROCS(4)
	src := flag.String("src", "", "master")
	dest := flag.String("dest", "", "slave")
	pullrate := flag.Int("pullrate", 400, "rdb pull rate in Mbits/s")
	buffer := flag.Int("buffer", 100, "buffer x10000 records")
	flag.Parse()

	if *pullrate < 100 {
		*pullrate = 100
	}
	if *buffer < 100 {
		*buffer = 100
	} else if *buffer > 1000 {
		*buffer = 1000
	}

	stdlog.Println("slaveof-proxy 1.0.1")
	stdlog.Printf("from [%s] to [%s]\n", *src, *dest)
	stdlog.Printf("pull [%d] buffer [%d]\n", *pullrate, *buffer)
	stdlog.Println("SYNC ...")

	client, err := NewClient(*src, *dest, *buffer)
	if err != nil {
		stdlog.Println("ERR", err)
		return
	}
	client.SetPullRate(*pullrate / 8 * 1024 * 1024)
	client.Sync()
}

func init() {
	// 全局日志前缀
	stdlog.SetPrefix(func() string {
		t := time.Now()
		return fmt.Sprintf("[%d-%02d-%02d %02d:%02d:%02d] ", t.Year(), t.Month(), t.Day(), t.Hour(), t.Minute(), t.Second())
	})
}

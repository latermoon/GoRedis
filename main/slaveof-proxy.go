package main

import (
	. "./tool/slaveof"
	"GoRedis/libs/stdlog"
	"flag"
	"fmt"
	"net"
	"runtime"
	"time"
)

// ./slaveof-proxy -src localhost:6379 -dest remote:6379 -pullrate 400
func main() {
	runtime.GOMAXPROCS(4)
	src := flag.String("src", "", "master")
	dest := flag.String("dest", "", "slave")
	pullrate := flag.Int("pullrate", 400, "rdb pull rate in Mbits/s")
	flag.Parse()

	srcConn, e1 := net.Dial("tcp", *src)
	if e1 != nil {
		stdlog.Println("connect master error", e1)
		return
	}
	destConn, e2 := net.Dial("tcp", *dest)
	if e2 != nil {
		stdlog.Println("connect slave error", e2)
		return
	}

	if *pullrate < 100 {
		*pullrate = 100
	}

	client := NewClient(srcConn, destConn)
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

package main

// 通过MONITOR将Redis指令转发到另外的Redis或GoRedis
// TODO 过程式代码优化
import (
	// . "GoRedis/goredis"
	"GoRedis/libs/shardredis"
	"GoRedis/libs/stdlog"
	"flag"
	"fmt"
	"runtime"
	"time"
)

func init() {
	runtime.GOMAXPROCS(4)

	stdlog.SetPrefix(func() string {
		t := time.Now()
		return fmt.Sprintf("[%d-%02d-%02d %02d:%02d:%02d] ", t.Year(), t.Month(), t.Day(), t.Hour(), t.Minute(), t.Second())
	})
}

func main() {
	src := flag.String("src", "", "source host")
	dest := flag.String("dest", "", "dest host")
	mode := flag.String("mode", "", "r/w/rw")
	flag.Parse()

	if len(*src) == 0 || len(*dest) == 0 {
		stdlog.Println("must set -src or -dest")
		return
	}
	if len(*mode) == 0 {
		stdlog.Println("must set -mode [r|w|rw]")
		return
	}

	monitor := shardredis.NewMonitorRedirect(*src, *dest, *mode)
	if err := monitor.Start(); err != nil {
		panic(err)
	}
}

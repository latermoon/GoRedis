package main

import (
	"GoRedis/goredis_proxy"
	"GoRedis/libs/stdlog"
	"flag"
	"fmt"
	"runtime"
	"time"
)

// go run redis-proxy.go -p 1603 -master localhost:6379 -slave localhost:6389
func main() {
	runtime.GOMAXPROCS(8)
	// options
	opt := goredis_proxy.NewOptions()
	// flags
	version := flag.Bool("v", false, "print goredis-proxy version")
	flag.StringVar(&opt.Host, "h", "", "server host")
	flag.IntVar(&opt.Port, "p", 1602, "server port")
	flag.StringVar(&opt.MasterHost, "master", "", "master")
	flag.StringVar(&opt.SlaveHost, "slave", "", "slave")
	flag.StringVar(&opt.Mode, "mode", "rw", "rw(default) = only read from slave; rrw = read from both")
	flag.IntVar(&opt.PoolSize, "poolsize", 100, "pool for remote server")
	flag.Parse()

	if *version {
		fmt.Println("redis-proxy ", goredis_proxy.VERSION)
		return
	}

	if len(opt.MasterHost) == 0 || len(opt.SlaveHost) == 0 {
		stdlog.Println("bad master/slave")
		return
	}

	stdlog.Println("redis-proxy ", goredis_proxy.VERSION)
	stdlog.Printf("master:[%s], slave:[%s]\n", opt.MasterHost, opt.SlaveHost)
	stdlog.Println("listen", opt.Addr())

	// start
	server := goredis_proxy.NewProxy(opt)
	err := server.Init()
	if err != nil {
		panic(err)
	}
	err = server.Listen(opt.Addr())
	if err != nil {
		panic(err)
	}
}

func init() {
	// 全局日志前缀
	stdlog.SetPrefix(func() string {
		t := time.Now()
		return fmt.Sprintf("[%d-%02d-%02d %02d:%02d:%02d] ", t.Year(), t.Month(), t.Day(), t.Hour(), t.Minute(), t.Second())
	})
}

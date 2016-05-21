// goredis-server启动函数
// @latermoon

package main

import (
	"../goredis_server"
	"GoRedis/libs/levelredis"
	"GoRedis/libs/stdlog"
	"flag"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"time"
)

// go run goredis-server.go -h localhost -p 1602
// go run goredis-server.go -procs 8 -p 17600
// go run goredis-server.go -slaveof localhost:1603
// go run goredis-server.go -dbpath /data/ -logpath /home/logs/
func main() {
	version := flag.Bool("v", false, "print version")
	host := flag.String("h", "0.0.0.0", "server host")
	port := flag.Int("p", 1602, "server port")
	slaveof := flag.String("slaveof", "", "replication")
	procs := flag.Int("procs", 8, "GOMAXPROCS, CPU")
	repair := flag.Bool("repair", false, "repair rocksdb")
	dbpath := flag.String("dbpath", "/data/", "rocksdb path, recommend use SSD")
	logpath := flag.String("logpath", "/data/", "all logs, include synclog,aof")
	flag.Parse()

	if *version {
		fmt.Println("goredis-server", goredis_server.VERSION)
		return
	}

	if !dirExist(*dbpath) {
		stdlog.Println("-dbpath", *dbpath, "not exist")
		return
	}
	if !dirExist(*logpath) {
		stdlog.Println("-logpath", *logpath, "not exist")
		return
	}

	runtime.GOMAXPROCS(*procs)

	// Options
	opt := goredis_server.NewOptions()
	opt.SetHost(*host)
	opt.SetPort(*port)
	opt.SetDBPath(joinGoRedisPath(*dbpath, *port))
	opt.SetLogPath(joinGoRedisPath(*logpath, *port))
	// ensure
	os.Mkdir(opt.DBPath(), os.ModePerm)
	os.Mkdir(opt.LogPath(), os.ModePerm)

	// split -slaveof host:port
	if len(*slaveof) > 0 {
		hostPort := strings.Split(*slaveof, ":")
		if len(hostPort) != 2 {
			panic("bad slaveof")
		}
		p, e := strconv.Atoi(hostPort[1])
		if e != nil {
			panic(e)
		}
		opt.SetSlaveOf(hostPort[0], p)
	}

	// 重定向日志输出位置
	if err := redirectStdout(opt.LogPath()); err != nil {
		panic(err)
	}

	// repair
	if *repair {
		dbhome := filepath.Join(opt.DBPath(), "db0")
		if !dirExist(dbhome) {
			stdlog.Println("db not exist")
		} else {
			stdlog.Println("start repair", dbhome)
			levelredis.Repair(dbhome)
			stdlog.Println("repair finish")
		}
		return
	}

	stdlog.Println("========================================")
	stdlog.Println("server init, version", goredis_server.VERSION, "...")
	stdlog.Printf("dbpath:%s, logpath:%s\n", opt.DBPath(), opt.LogPath())

	// GoRedis Server
	server := goredis_server.NewGoRedisServer(opt)
	if err := server.Init(); err != nil {
		panic(err)
	}
	if err := server.Listen(); err != nil {
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

// 将stdout, stderr重定向到指定文件
func redirectStdout(logpath string) (err error) {
	// stdout
	oldout := os.Stdout
	if os.Stdout, err = os.OpenFile(filepath.Join(logpath, "stdout.log"), os.O_WRONLY|os.O_CREATE|os.O_APPEND, os.ModePerm); err != nil {
		return
	}
	stdlog.SetOutput(io.MultiWriter(oldout, os.Stdout))

	// stderr
	if os.Stderr, err = os.OpenFile(filepath.Join(logpath, "stderr.log"), os.O_WRONLY|os.O_CREATE|os.O_APPEND, os.ModePerm); err != nil {
		return
	}
	return
}

// GoRedis文件夹路径，返回格式如：/data/goredis_1602/
func joinGoRedisPath(path string, port int) string {
	return filepath.Join(path, fmt.Sprintf("goredis_%d/", port))
}

// 检查目录是否存在
func dirExist(dir string) bool {
	info, err := os.Stat(dir)
	if err == nil {
		return info.IsDir()
	} else {
		return !os.IsNotExist(err) && info.IsDir()
	}
}

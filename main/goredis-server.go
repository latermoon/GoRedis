// goredis-server启动函数
// @latermoon

package main

import (
	"GoRedis/goredis_server"
	"GoRedis/libs/levelredis"
	"GoRedis/libs/stdlog"
	"errors"
	"flag"
	"fmt"
	"io"
	"os"
	"runtime"
	"strconv"
	"strings"
	"time"
)

// go run goredis-server.go -h localhost -p 1602
// go run goredis-server.go -procs 8 -p 17600
// go run goredis-server.go -slaveof localhost:1603
func main() {
	version := flag.Bool("v", false, "print goredis-server version")
	host := flag.String("h", "", "server host")
	port := flag.Int("p", 1602, "server port")
	slaveof := flag.String("slaveof", "", "replication")
	procs := flag.Int("procs", 8, "GOMAXPROCS")
	repair := flag.Bool("repair", false, "repaire rocksdb")
	datapath := flag.String("datapath", "/data", "config goredis data path default path [/data/goredis_${port}/]")
	logpath := flag.String("logpath", "/home/logs/", "config goredis log path and synclog path ,default path [/home/logs/goredis_${port}]")
	flag.Parse()

	if *version {
		fmt.Println("goredis-server", goredis_server.VERSION)
		return
	}

	runtime.GOMAXPROCS(*procs)

	opt := goredis_server.NewOptions()
	opt.SetBind(fmt.Sprintf("%s:%d", *host, *port))
	dbhome := dbHome(*datapath, *port)
	opt.SetDirectory(dbhome)

	if len(*slaveof) > 0 {
		h, p, e := splitHostPort(*slaveof)
		if e != nil {
			panic(e)
		}
		opt.SetSlaveOf(h, p)
	}

	// 重定向日志输出位置
	logdir := redirectLogOutput(*logpath, *port)
	opt.SetLogDir(logdir)
	stdlog.Println("logdir:[" + logdir + "]\tdbhome:[" + dbhome + "]")

	// repair
	if *repair {
		dbhome := opt.Directory() + "db0"
		finfo, e1 := os.Stat(dbhome)
		if os.IsNotExist(e1) || !finfo.IsDir() {
			stdlog.Println("db not exist")
			return
		} else {
			stdlog.Println("start repair", dbhome)
			levelredis.Repair(dbhome)
			stdlog.Println("repair finish")
		}
		return
	}

	stdlog.Println("========================================")
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

/**
 * 构建dbhome，使用datapath中的路径
 */
func dbHome(datapath string, port int) string {

	dbhome := fmt.Sprintf("%s/goredis_%d/", datapath, port)

	finfo, err := os.Stat(dbhome)
	if os.IsNotExist(err) || !finfo.IsDir() {
		os.MkdirAll(dbhome, os.ModePerm)
	}
	return dbhome
}

/**
 * 将Stdout, Stderr重定向到指定文件
 * 并返回当前日志路径
 */
func redirectLogOutput(directory string, port int) string {

	oldout := os.Stdout

	logpath := fmt.Sprintf("%s/goredis_%d/", directory, port)

	loginfo, err := os.Stat(logpath)

	/**
	 * 如果logpath不存在
	 * 则创建
	 */
	if os.IsNotExist(err) || !loginfo.IsDir() {
		os.MkdirAll(logpath, os.ModePerm)
	}

	os.Stdout, err = os.OpenFile(logpath+"stdout.log", os.O_WRONLY|os.O_CREATE|os.O_APPEND, os.ModePerm)

	if err != nil {
		panic(err)
	}
	// 同时输出到屏幕和文件
	stdlog.SetOutput(io.MultiWriter(oldout, os.Stdout))

	os.Stderr, err = os.OpenFile(logpath+"stderr.log", os.O_WRONLY|os.O_CREATE|os.O_APPEND, os.ModePerm)
	if err != nil {
		panic(err)
	}

	return logpath
}

func splitHostPort(addr string) (host string, port int, err error) {
	tmp := strings.Split(addr, ":")
	if len(tmp) != 2 {
		err = errors.New("bad addr:" + addr)
		return
	}
	host = tmp[0]
	port, err = strconv.Atoi(tmp[1])
	return
}

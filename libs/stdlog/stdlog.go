package stdlog

/*
一个简洁通用的日志，模仿golang原生的log
【最主要的变化】
最主要的变化是使用函数作为Prefix，可以灵活输出想要的时间格式，以及Target、runtime.Caller等，而golang原生log无法定制
同时只保留最主要的Print/Printf/Println函数，清晰易读

Install:
go get github.com/latermoon/GoRedis/libs/stdlog

import "github.com/latermoon/GoRedis/libs/stdlog"

一、单例用法

stdlog.Println("init ...")

// （可选）设置函数前缀
stdlog.SetPrefix(func() string {
	t := time.Now()
	return fmt.Sprintf("[%d-%02d-%02d %02d:%02d:%02d] ", t.Year(), t.Month(), t.Day(), t.Hour(), t.Minute(), t.Second())
})

Output:
	[2014-01-04 17:15:49] init ...

二、实例用法

1. 获取指定的Logger，如果不做任何配置，用起来和stdlog单例一模一样
var log2 = stdlog.Log("log2")

2. 配置输出路径
out, _ := os.OpenFile("/tmp/error.log", os.O_WRONLY|os.O_CREATE|os.O_APPEND, os.ModePerm)
log2.SetOutput(out)

3. 配置额外的前缀
log2.SetPrefix(func() string {
	return stdlog.Prefix() + "[Target] "
})

4. Output:
	[2014-01-04 17:15:49] [Target] init ...

*/
import (
	"io"
	"os"
	"sync"
)

var (
	stdlogger    Logger            // 默认Logger
	caches       map[string]Logger // 缓存
	mu           sync.Mutex
	globalPrefix func() string // 默认前缀函数
	globalOut    io.Writer     // 默认输出位置
)

func init() {
	caches = make(map[string]Logger)
	globalOut = os.Stdout // 默认输出
	stdlogger = Log("")   // 默认日志
}

/**
 * 获取一个指定的Logger
 * @param name 日志名称
 */
func Log(name string) (l Logger) {
	mu.Lock()
	defer mu.Unlock()
	var ok bool
	l, ok = caches[name]
	if !ok {
		l = &SimpleLogger{}
		caches[name] = l
	}
	return
}

func SetPrefix(fn func() string) {
	globalPrefix = fn
}

func Prefix() func() string {
	return globalPrefix
}

func SetOutput(w io.Writer) {
	globalOut = w
}

func Output() io.Writer {
	return globalOut
}

func Println(v ...interface{}) {
	stdlogger.Println(v...)
}

func Printf(format string, v ...interface{}) {
	stdlogger.Printf(format, v...)
}

func Print(v ...interface{}) {
	stdlogger.Print(v...)
}

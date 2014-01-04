package stdlog

/*
一个简洁通用的日志，模仿golang原生的log，最主要变化是使用函数作为Prefix，同时提供工厂方法获取Logger
使用函数Prefix，可以灵活输出想要的时间格式，以及Target、runtime.Caller等，而原生log完全无法定制
同时只保留Print/Printf/Println函数，清晰易读

Install:
go get github.com/latermoon/GoRedis/libs/stdlog

import "xxx/stdlog"

// 设置函数前缀（可选）
stdlog.SetPrefix(func() string {
	t := time.Now()
	return fmt.Sprintf("[%d-%02d-%02d %02d:%02d:%02d] ", t.Year(), t.Month(), t.Day(), t.Hour(), t.Minute(), t.Second())
})

// 默认输出到os.Stdout
stdlog.Println("init ...")
Output:
	[2014-01-04 17:15:49] init ...

也可以获取多个Logger实例，配置输出路径
var logger = stdlog.Log("error")

func init() {
	out, _ := os.OpenFile("/tmp/error.log", os.O_WRONLY|os.O_CREATE|os.O_APPEND, os.ModePerm)
	logger.SetOut(out)
}

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

func SetOutput(w io.Writer) {
	globalOut = w
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

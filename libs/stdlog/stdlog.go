// Copyright 2013 Latermoon. All rights reserved.

// 一个简洁通用的log工具，参照golang原生的log，增加灵活的用法
//
// 最主要的变化是使用函数作为Prefix，可以灵活输出想要的时间格式，以及Target、runtime.Caller等，而golang原生log无法定制
//
// 同时只保留最主要的Print/Printf/Println函数
package stdlog

/*
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
	caches        map[string]Logger // 缓存
	mu            sync.Mutex
	defaultLogger Logger        // 默认Logger
	defaultPrefix func() string // 默认前缀函数
	defaultOutput io.Writer     // 默认输出位置
)

func init() {
	caches = make(map[string]Logger)
	defaultOutput = os.Stdout // 默认输出os.Stdout
	defaultLogger = Log("")
}

// 获取一个指定的Logger
// name
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
	defaultPrefix = fn
}

func Prefix() func() string {
	return defaultPrefix
}

func SetOutput(w io.Writer) {
	defaultOutput = w
}

func Output() io.Writer {
	return defaultOutput
}

func Println(v ...interface{}) {
	defaultLogger.Println(v...)
}

func Printf(format string, v ...interface{}) {
	defaultLogger.Printf(format, v...)
}

func Print(v ...interface{}) {
	defaultLogger.Print(v...)
}

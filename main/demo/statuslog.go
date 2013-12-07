package main

import (
	// "../../libs/statlog"
	"github.com/latermoon/GoRedis/libs/statlog"
	"os"
	"runtime"
)

func main() {
	memstats()
}

func demo1() {
	slog := statlog.NewStatLogger(os.Stdout)
	opt := &statlog.Opt{Padding: 8}

	slog.Add(statlog.TimeItem("time"))
	slog.Add(statlog.Item("total", func() interface{} {
		return "10"
	}, opt))
	slog.Add(statlog.Item("buffer", func() interface{} {
		return 10342
	}, opt))

	slog.Start()
}

func memstats() {
	l := statlog.NewStatLogger(os.Stdout)
	opt := &statlog.Opt{Padding: 10}

	var m runtime.MemStats
	l.BeforePrint(func() {
		go func() {
			runtime.ReadMemStats(&m)
		}()
	})

	l.Add(statlog.TimeItem("time"))
	// 程序向操作系统请求的内存的字节数
	l.Add(statlog.Item("HeapSys", func() interface{} {
		return m.HeapSys
	}, opt))
	// 当前堆中已经分配的字节数
	l.Add(statlog.Item("HeapAlloc", func() interface{} {
		return m.HeapAlloc
	}, opt))
	// 堆中未使用的字节数
	l.Add(statlog.Item("HeapIdle", func() interface{} {
		return m.HeapIdle
	}, opt))
	// 归还给操作系统的字节数
	l.Add(statlog.Item("HeapReleased", func() interface{} {
		return m.HeapReleased
	}, opt))

	l.Start()
}

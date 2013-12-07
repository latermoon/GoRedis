package main

import (
	"../../libs/statlog"
	// "fmt"
	"os"
	"runtime"
)

func main() {
	memstats()
}

func demo1() {
	l := statlog.NewStatLogger(os.Stdout)
	opt := &statlog.Opt{Padding: 8}

	l.Add(statlog.TimeItem("time"))
	l.Add(statlog.Item("total", func() interface{} {
		return "10"
	}, opt))
	l.Add(statlog.Item("buffer", func() interface{} {
		return 10342
	}, opt))

	l.Start()
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

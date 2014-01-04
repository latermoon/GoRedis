package main

import (
	"../../libs/stdlog"
	"fmt"
	"time"
	// "os"
)

var log = stdlog.Log("redis")

func init() {
	// 设置每行日志的前缀
	stdlog.SetPrefix(func() string {
		t := time.Now()
		return fmt.Sprintf("[%d-%02d-%02d %02d:%02d:%02d] ", t.Year(), t.Month(), t.Day(), t.Hour(), t.Minute(), t.Second())
	})
}

func main() {
	log.Printf("[%s]\n", "192.168.1.101")
	log.Println("ca!")
}

package main

import (
	"../../libs/iotool"
	// "bufio"
	"github.com/latermoon/GoRedis/libs/stdlog"
	"os"
)

func main() {
	src := "/Users/latermoon/Downloads/mercedes-benz-e-class-coupe-c207_wallpaper_05_1920x1200_03-2013.jpg"
	dst := "/tmp/mercedes-benz-e-class-coupe-c207_wallpaper_05_1920x1200_03-2013.jpg"
	// src := "..."
	// dst := "..."
	srcfile, e1 := os.OpenFile(src, os.O_RDONLY, os.ModePerm)
	dstfile, e2 := os.OpenFile(dst, os.O_WRONLY|os.O_CREATE, os.ModePerm)
	if e1 != nil || e2 != nil {
		stdlog.Println(e1, e2)
	}
	n, err := iotool.RateLimitCopy(dstfile, srcfile, 100*1024*1024, func(written int64, rate int) {
		stdlog.Println("copy:", written, "rate:", rate)
	})
	stdlog.Println(n, err)
}

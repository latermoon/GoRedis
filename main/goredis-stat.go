package main

/*
goredis-stat 输出GoRedis实时状态
通过获取info里的内容，构造定时输出
*/

import (
	"./tool/stat"
	"flag"
	"fmt"
	"runtime"
	"strings"
)

//go run goredis-stat.go -info memory -field "m_Alloc=Alloc;m_Mallocs=Mallocs;m_Frees=Frees;m_HeapAlloc=HeapAlloc;m_HeapIdle=HeapIdle;m_HeapReleased=HeapReleased;m_HeapObjects=HeapObjects" -h goredis-nearby-a001 -p 18400
//go run goredis-stat.go -info memory -field "m_Alloc=Alloc;m_Mallocs=Mallocs;m_Frees=Frees;m_HeapAlloc=HeapAlloc;m_HeapIdle=HeapIdle;m_HeapReleased=HeapReleased;m_HeapObjects=HeapObjects" -p 1602
func main() {
	version := flag.Bool("v", false, "print goredis-stat version")
	hostPtr := flag.String("h", "", "host")
	portPtr := flag.Int("p", 1602, "port")
	infoPtr := flag.String("info", "", "info")
	fieldPtr := flag.String("field", "", "field list")
	procsPtr := flag.Int("procs", 2, "GOMAXPROCS")
	flag.Parse()

	if *version {
		fmt.Println("goredis-stat", "0.0.1")
		return
	}

	if len(*infoPtr) == 0 {
		fmt.Println("miss -info [section]")
		return
	}
	if len(*fieldPtr) == 0 {
		fmt.Println("miss -field [field=name;field=name;...]")
		return
	}

	runtime.GOMAXPROCS(*procsPtr)

	host := fmt.Sprintf("%s:%d", *hostPtr, *portPtr)
	client := stat.NewStatClient(host, *infoPtr)

	fields := make([]string, 0, 100)

	pairs := strings.Split(*fieldPtr, ";")
	for i := 0; i < len(pairs); i++ {
		if len(pairs[i]) == 0 {
			continue
		}
		kv := strings.Split(pairs[i], "=")
		if len(kv) != 2 {
			continue
		} else if len(kv[0]) == 0 || len(kv[1]) == 0 {
			continue
		}
		fields = append(fields, kv[0])
		fields = append(fields, kv[1])
	}

	// fmt.Println("fields:", fields)
	client.SetFields(fields...)

	client.Connect()
}

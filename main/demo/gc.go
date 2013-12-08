package main

import (
	"fmt"
	"runtime"
	// "runtime/debug"
	"os"
	"runtime/pprof"
)

func main() {
	pprof.StartCPUProfile(os.Stdout)
	a := "name"
	m := map[string]interface{}{"name": "latermoon", "age": 12}
	fmt.Println(runtime.NumGoroutine(), a, m)
	pprof.StopCPUProfile()
}

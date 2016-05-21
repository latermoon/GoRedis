package main

import (
	"GoRedis/libs/stat"
	// "fmt"
	"math/rand"
	"os"
	"time"
)

func main() {
	var recv int64 = 10
	go func() {
		for {
			recv += rand.Int63n(1000)
			time.Sleep(time.Millisecond * 300)
		}
	}()

	s := stat.New(os.Stdout)
	go func() {
		time.Sleep(time.Second * 10)
		s.Close()
	}()
	s.Add(stat.TextItem("time", 8, func() interface{} { return stat.TimeString() }))
	s.Add(stat.IncrItem("recv", 8, func() int64 { return recv }))
	// s.Add(stat.NewItem("time", 8, "OK"))
	// s.Add(stat.NewItem("time", 8, "OK"))
	s.Start()
}

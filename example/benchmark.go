package main

import (
	//"encoding/json"
	"fmt"
	"github.com/garyburd/redigo/redis"
	"time"
)

//var profileJson = "501f8365ccd569a138000000501f8365ccd569a138000000501f8365ccd569a138000000501f8365ccd569a138000000501f8365ccd569a138000000501f8365ccd569a138000000501f8365ccd569a138000000501f8365ccd569a138000000501f8365ccd569a138000000501f8365ccd569a138000000501f8365ccd569a138000000501f8365ccd569a138000000501f8365ccd569a138000000501f8365ccd569a138000000501f8365ccd569a138000000501f8365ccd569a138000000501f8365ccd569a138000000501f8365ccd569a138000000501f8365ccd569a138000000501f8365ccd569a138000000501f8365ccd569a138000000501f8365ccd569a138000000501f8365ccd569a138000000501f8365ccd569a138000000501f8365ccd569a138000000501f8365ccd569a138000000501f8365ccd569a138000000501f8365ccd569a138000000501f8365ccd569a138000000501f8365ccd569a138000000501f8365ccd569a138000000"

func thread(conn redis.Conn, count int, ch chan int) {
	t1 := time.Now()
	for i := 0; i < count; i++ {
		conn.Do("GET", "name")
	}
	ch <- 1
	t2 := time.Now()
	fmt.Println("Done in:", t2.Sub(t1))
}

func main() {
	//host := ":6379"
	host := ":1603"

	chanCount := 10
	countPerThread := 10000
	clients := make([]redis.Conn, chanCount)
	ch := make(chan int, chanCount)
	for i := 0; i < chanCount; i++ {
		clients[i], _ = redis.Dial("tcp", host)
	}
	fmt.Println("start...")
	t1 := time.Now()
	for i := 0; i < chanCount; i++ {
		go thread(clients[i], countPerThread, ch)
	}
	for i := 0; i < chanCount; i++ {
		<-ch
	}
	elapsed := time.Now().Sub(t1)
	qps := float64(chanCount*countPerThread) / elapsed.Seconds()
	fmt.Println("count:", chanCount*countPerThread, "elapsed:", elapsed, "qps:", qps)
}

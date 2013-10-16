package main

import (
	"fmt"
	"github.com/garyburd/redigo/redis"
	"math/rand"
	"strconv"
	"time"
)

func thread(conn redis.Conn, count int, ch chan int) {
	t1 := time.Now()
	for i := 0; i < count; i++ {
		rndid := 20000000 + rand.Intn(2000000)*10
		//conn.Do("GET", "user:"+strconv.Itoa(rndid)+":sex")
		conn.Do("SET", "user:"+strconv.Itoa(rndid)+":sex_f_m", "FM..FM..FM..")
		// if e1 == nil {
		// 	if reply != nil {
		// 		fmt.Println(string(reply.([]byte)))
		// 	}
		// }
	}
	ch <- 1
	t2 := time.Now()
	fmt.Println("Done in:", t2.Sub(t1))
}

func main() {
	rand.Seed(time.Now().UnixNano())

	//host := ":6379"
	host := ":1603"

	chanCount := 100
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

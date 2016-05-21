package main

import (
	"fmt"
	"github.com/latermoon/redigo/redis"
	"math/rand"
	"strconv"
	"sync"
	"time"
)

var i int = 0
var mu sync.Mutex

func incrCounter() {
	mu.Lock()
	defer mu.Unlock()
	i++
	if i%100000 == 0 {
		fmt.Println("access count:", i)
	}
}

func thread(conn redis.Conn, count int, ch chan int) {
	t1 := time.Now()

	for i := 0; i < count; i++ {
		rndid := 1000000 + rand.Intn(75000000)
		remoteid := strconv.Itoa(rndid)
		conn.Do("GET", "user:"+remoteid+":profile")
		// conn.Do("GET", "user:"+remoteid+":password")
		// conn.Do("GET", "user:"+remoteid+":cflag")
		// conn.Do("GET", "user:"+remoteid+":sessionid")
		// conn.Do("GET", "user:"+remoteid+":uid")
		// num := 20000000 + rand.Intn(2000000)*10 + rand.Intn(4)
		// reply, _ := conn.Do("aof_push_async", "user:"+strconv.Itoa(rndid)+":history", strconv.Itoa(num))
		// if reply == nil {
		// 	fmt.Println(reply)
		// }

		//conn.Do("SET", "user:"+strconv.Itoa(rndid)+":sex_f_m", "FM..FM..FM..")
		// if e1 == nil {
		// 	if reply != nil {
		// 		fmt.Println(string(reply.([]byte)))
		// 	}
		// }
		incrCounter()
	}
	ch <- 1
	t2 := time.Now()
	fmt.Println("Done in:", t2.Sub(t1))
}

func main() {
	rand.Seed(time.Now().UnixNano())

	//host := ":6379"
	host := ":17600"

	chanCount := 50
	countPerThread := 2000000 * 10
	clients := make([]redis.Conn, chanCount)
	ch := make(chan int, chanCount)
	for i := 0; i < chanCount; i++ {
		var err error
		clients[i], err = redis.Dial("tcp", host)
		if err != nil {
			panic(err)
		}
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

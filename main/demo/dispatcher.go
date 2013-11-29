package main

import (
	qp "../../goredis_server/libs/queueprocess"
	"fmt"
)

func main() {
	queue := qp.NewQueueProcess(10, func(t qp.Task) {
		fmt.Println("task", t)
	})
	for i := 0; i < 1000; i++ {
		queue.Process(i, fmt.Sprintf("name %d", i))
	}
	ch := make(chan int, 0)
	ch <- 0

}

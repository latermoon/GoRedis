package queueprocess

import (
	"container/list"
	"fmt"
	"sync"
	"time"
)

// 要处理的队列元素
type Task interface{}

// 处理函数
type Handler func(t Task)

// 队列管理器
type QueueProcess struct {
	thread     int
	handler    Handler
	mutexs     []*sync.Mutex
	queues     []*list.List
	shouldStop bool
}

func NewQueueProcess(thread int, handler Handler) (q *QueueProcess) {
	q = &QueueProcess{}
	q.thread = thread
	q.handler = handler
	q.queues = make([]*list.List, thread)
	q.mutexs = make([]*sync.Mutex, thread)
	for i := 0; i < thread; i++ {
		q.queues[i] = list.New()
		q.mutexs[i] = &sync.Mutex{}
		go q.processQueue(i)
	}
	return
}

func (q *QueueProcess) processQueue(i int) {
	queue := q.queues[i]
	mu := q.mutexs[i]
	sleepCount := 0
	for {
		if q.shouldStop {
			break
		}
		// LPop
		mu.Lock()
		elem := queue.Front()
		if queue.Len() > 200 {
			fmt.Println(elem.Value)
		}
		if elem == nil {
			mu.Unlock()
			sleepCount++
			if sleepCount > 100 {
				sleepCount = 100
			}
			time.Sleep(time.Millisecond * time.Duration(10*sleepCount))
			continue
		}
		sleepCount = 0
		queue.Remove(elem)
		mu.Unlock()
		// Process
		q.handler(elem.Value.(Task))
	}
}

func (q *QueueProcess) Process(hash int, t Task) {
	if q.shouldStop {
		panic("queue already stop")
	}
	idx := hash % q.thread
	if idx < 0 {
		panic("hash must > 0")
	}
	q.mutexs[idx].Lock()
	defer q.mutexs[idx].Unlock()
	q.queues[idx].PushBack(t)
}

func (q *QueueProcess) Stop() {
	q.shouldStop = true
}

func (q *QueueProcess) QueueLen() (ns []int) {
	ns = make([]int, 0, q.thread)
	for i := 0; i < q.thread; i++ {
		ns = append(ns, q.queues[i].Len())
	}
	return
}

func (q *QueueProcess) Len() (n int) {
	for i := 0; i < q.thread; i++ {
		n += q.queues[i].Len()
	}
	return
}

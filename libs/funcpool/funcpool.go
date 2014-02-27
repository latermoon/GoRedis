package funcpool

import (
	"math/rand"
	"sync"
	"time"
)

// 线程池(goroutine pool)
type FuncPool struct {
	size  int
	tasks []chan func()
	mu    []*sync.Mutex
	wg    *sync.WaitGroup
}

func New(size int) (f *FuncPool) {
	f = &FuncPool{
		size:  size,
		tasks: make([]chan func(), size),
		mu:    make([]*sync.Mutex, size),
		wg:    &sync.WaitGroup{},
	}

	rand.Seed(time.Now().UnixNano())

	for i := 0; i < size; i++ {
		f.tasks[i] = make(chan func())
		f.mu[i] = &sync.Mutex{}
		go f.runloop(i)
	}

	return
}

func (f *FuncPool) runloop(i int) {
	for {
		fn, ok := <-f.tasks[i]
		if !ok {
			break
		}
		fn()
		f.wg.Done()
	}
}

// 指定运行队列，如果hash<0，将使用随机队列
func (f *FuncPool) Run(hash int, fn func()) int {
	f.wg.Add(1)
	i := hash % len(f.mu)
	if i < 0 {
		i = rand.Intn(f.size)
	}
	f.tasks[i] <- fn
	return i
}

func (f *FuncPool) Wait() {
	f.wg.Wait()
}

func (f *FuncPool) Close() {
	f.wg.Wait()
	for i := 0; i < f.size; i++ {
		close(f.tasks[i])
	}
}

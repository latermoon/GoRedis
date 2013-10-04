package monitor

import (
	"sync"
)

// 包装一组Counter，比提供简化的获取函数
type Counters struct {
	table map[string]*Counter
	mutex sync.Mutex
}

func NewCounters() (c *Counters) {
	c = &Counters{}
	c.table = make(map[string]*Counter)
	return
}

func (c *Counters) Len() int {
	return len(c.table)
}

// 获取并自动创建
func (c *Counters) Get(name string) (counter *Counter) {
	c.mutex.Lock()
	defer c.mutex.Unlock()
	var exist bool
	counter, exist = c.table[name]
	if !exist {
		counter = NewCounter()
		c.table[name] = counter
	}
	return
}

func (c *Counters) Counters() (counters []*Counter) {
	counters = make([]*Counter, 0, len(c.table))
	for _, val := range c.table {
		counters = append(counters, val)
	}
	return
}

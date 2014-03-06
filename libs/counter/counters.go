package counter

import (
	"sync"
)

// 包装一组Counter，比提供简化的获取函数
type Counters struct {
	table map[string]*Counter
	mu    sync.Mutex
}

func NewCounters() (c *Counters) {
	c = &Counters{
		table: make(map[string]*Counter),
	}
	return
}

func (c *Counters) Len() int {
	return len(c.table)
}

// 获取并自动创建
func (c *Counters) Get(name string) *Counter {
	counter, exist := c.table[name]
	if !exist {
		c.mu.Lock()
		counter, exist = c.table[name]
		if !exist {
			counter = New(0)
			c.table[name] = counter
		}
		c.mu.Unlock()
	}
	return counter
}

func (c *Counters) Names() (names []string) {
	c.mu.Lock()
	defer c.mu.Unlock()

	names = make([]string, 0, len(c.table))
	for key, _ := range c.table {
		names = append(names, key)
	}
	return
}

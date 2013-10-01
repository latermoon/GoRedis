package monitor

import (
	"sync"
)

// 计数器
type Counter struct {
	count     int
	prevCount int
	mutex     sync.Mutex
}

func NewCounter() (c *Counter) {
	c = &Counter{}
	c.count = 0
	c.prevCount = 0
	return
}

func (c *Counter) Count() int {
	return c.count
}

func (c *Counter) Incr(i int) {
	c.mutex.Lock()
	c.count += i
	c.mutex.Unlock()
}

func (c *Counter) Decr(i int) {
	c.mutex.Lock()
	c.count -= i
	c.mutex.Unlock()
}

func (c *Counter) Clear() {
	c.mutex.Lock()
	c.count = 0
	c.prevCount = 0
	c.mutex.Unlock()
}

func (c *Counter) ChangedCount() (chg int) {
	c.mutex.Lock()
	chg, c.prevCount = c.count-c.prevCount, c.count
	c.mutex.Unlock()
	return
}

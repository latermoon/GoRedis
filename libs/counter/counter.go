package counter

import (
	"sync"
)

// 计数器
type Counter struct {
	count     int64
	prevCount int64
	mutex     sync.Mutex
}

func NewCounter() (c *Counter) {
	c = &Counter{}
	c.count = 0
	c.prevCount = 0
	return
}

func (c *Counter) SetCount(i int64) {
	c.mutex.Lock()
	c.prevCount, c.count = c.count, i
	c.mutex.Unlock()
}

func (c *Counter) Count() int64 {
	return c.count
}

func (c *Counter) Incr(i int64) {
	c.mutex.Lock()
	c.count += i
	c.mutex.Unlock()
}

func (c *Counter) Decr(i int64) {
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

func (c *Counter) ChangedCount() (chg int64) {
	c.mutex.Lock()
	chg, c.prevCount = c.count-c.prevCount, c.count
	c.mutex.Unlock()
	return
}

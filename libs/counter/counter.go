package counter

import (
	"fmt"
	"sync"
	"sync/atomic"
)

// Atomic Counter, simple enough, easy to Incr/Decr
// c := Counter(0)
// c.SetCount(100)
// c.Incr(1) or c.Decr(1)
// fmt.Println(c) or c.Count()
type Counter int64

func (c *Counter) SetCount(val int64) {
	atomic.StoreInt64((*int64)(c), val)
}

func (c *Counter) Count() int64 {
	return atomic.LoadInt64((*int64)(c))
}

func (c *Counter) Incr(delta int64) int64 {
	return atomic.AddInt64((*int64)(c), delta)
}

func (c *Counter) Decr(delta int64) int64 {
	return atomic.AddInt64((*int64)(c), delta*-1)
}

func (c *Counter) String() string {
	return fmt.Sprint(c.Count())
}

// Counter Collection
// factory := NewFactory()
// factory.Get("set").Incr(1)
// factory.Get("get").Incr(1)
// factory.Get("del").Incr(1)
// factory.Get("total").Incr(3)
type Counters struct {
	table map[string]*Counter
	mu    sync.Mutex
}

func NewCounters() *Counters {
	return &Counters{
		table: make(map[string]*Counter),
	}
}

// Get or auto create a Counter by name
func (f *Counters) C(name string) *Counter {
	var c *Counter
	var ok bool
	if c, ok = f.table[name]; !ok {
		f.mu.Lock()
		if c, ok = f.table[name]; !ok {
			tmp := Counter(0)
			c = &tmp
			f.table[name] = c
		}
		f.mu.Unlock()
	}
	return c
}

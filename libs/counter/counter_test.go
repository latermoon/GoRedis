package counter

import (
	"github.com/facebookgo/ensure"
	"testing"
)

func TestCounter(t *testing.T) {
	c := Counter(0)
	c.Incr(1)
	c.Decr(2)
	ensure.True(t, c.Count() == -1)
}

func TestCounters(t *testing.T) {
	coll := NewCounters()
	coll.C("get").Incr(1)
	coll.C("set").Decr(-1)
	ensure.True(t, coll.C("get").Count() == 1)
	ensure.True(t, coll.C("del").Count() == 0)
}

package rocks

import (
	"math"
	"testing"
)

func TestIntAndBytes(t *testing.T) {
	t.Log(Int64ToBytes(-2560))
	t.Log(Int64ToBytes(-256))
	t.Log(Int64ToBytes(-1))
	t.Log(Int64ToBytes(0))
	t.Log(Int64ToBytes(1))
	t.Log(Int64ToBytes(256))
	t.Log("math.MaxInt64:", math.MaxInt64)
}

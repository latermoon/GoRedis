package rocks

import (
	"math"
	"testing"
)

func ScanAll(t *testing.T, db *DB) {
	min := []byte{0}
	max := []byte{math.MaxUint8}
	db.RangeEnumerate(min, max, IterForward, func(i int, key, value []byte, quit *bool) {
		t.Log(i, string(key), string(value))
	})
}

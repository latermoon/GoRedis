package rocks

import (
	"github.com/facebookgo/ensure"
	"testing"
)

func TestHash(t *testing.T) {
	db := New(newRocksDB(t))
	defer db.Close()

	h := db.Hash([]byte("user"))
	ensure.Nil(t, h.Set([]byte("name"), []byte("latermoon")))
	ensure.Nil(t, h.Set([]byte("age"), []byte("28")))
	ensure.Nil(t, h.Set([]byte("sex"), []byte("Male")))

	val, err := h.Get([]byte("name"))
	ensure.Nil(t, err)
	ensure.DeepEqual(t, val, []byte("latermoon"))

	val, err = h.Get([]byte("age"))
	ensure.Nil(t, err)
	ensure.DeepEqual(t, val, []byte("28"))

	val, err = h.Get([]byte("sex"))
	ensure.Nil(t, err)
	ensure.DeepEqual(t, val, []byte("Male"))

	// Exist and remove
	exist, err := h.Exist([]byte("age"))
	ensure.Nil(t, err)
	ensure.True(t, exist)

	err = h.Remove([]byte("name"))
	ensure.Nil(t, err)

	exist, err = h.Exist([]byte("name"))
	ensure.Nil(t, err)
	ensure.False(t, exist)

	vals, err := h.MGet([]byte("age"), []byte("sex"))
	ensure.Nil(t, err)
	ensure.True(t, len(vals) == 2)
	ensure.DeepEqual(t, vals[0], []byte("28"), vals[1], []byte("Male"))

	h.drop()

	db.RangeEnumerate([]byte{0}, []byte{'z'}, IterForward, func(i int, key, value []byte, quit *bool) {
		t.Log(i, string(key), string(value))
		t.Fail()
	})
}

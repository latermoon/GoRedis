package rocks

import (
	"github.com/facebookgo/ensure"
	"testing"
)

func TestList(t *testing.T) {
	db := New(newRocksDB(t))
	defer db.Close()

	l := db.List([]byte("list"))
	ensure.True(t, l.Len() == 0)

	l.RPush([]byte("a"), []byte("b"))

	l.LPush([]byte("-a"), []byte("-b"), []byte("-c"))

	ensure.True(t, l.Len() == 5)

	val, err := l.RPop()
	ensure.Nil(t, err)
	ensure.DeepEqual(t, val, []byte("b"))
	ensure.True(t, l.Len() == 4)

	val, err = l.LPop()
	ensure.Nil(t, err)
	ensure.DeepEqual(t, val, []byte("-c"))
	ensure.True(t, l.Len() == 3)

	val, err = l.Index(0)
	ensure.Nil(t, err)
	ensure.DeepEqual(t, val, []byte("-b"))

	val, err = l.Index(2)
	ensure.Nil(t, err)
	ensure.DeepEqual(t, val, []byte("a"))

	l.Range(0, -1, func(i int, value []byte, quit *bool) {
		t.Log(i, string(value))
	})

	l.RPop()
	l.RPop()
	l.RPop()

	val, err = l.RPop()
	ensure.Nil(t, err)
	ensure.True(t, val == nil)

	val, err = l.Index(30)
	ensure.Nil(t, err)
	ensure.True(t, val == nil)

	l.RPush([]byte("1"), []byte("2"), []byte("3"), []byte("4"))
	l.drop()
	ScanAll(t, db)
}

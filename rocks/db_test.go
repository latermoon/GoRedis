package rocks

import (
	"github.com/facebookgo/ensure"
	"github.com/tecbot/gorocksdb"
	"io/ioutil"
	"testing"
)

func TestOpenDB(t *testing.T) {
	db := New(newTestDB(t))
	defer db.Close()
}

func TestDBGetSet(t *testing.T) {
	db := New(newTestDB(t))
	defer db.Close()

	var (
		key   = []byte("name")
		value = []byte("latermoon")
	)

	ensure.Nil(t, db.RawSet(key, value))

	value, err := db.RawGet(key)
	ensure.Nil(t, err)
	ensure.DeepEqual(t, value, value)

	ensure.Nil(t, db.RawDelete(key))

}

func newTestDB(t *testing.T) *gorocksdb.DB {
	dir, err := ioutil.TempDir("", "rocks")
	ensure.Nil(t, err)

	opts := gorocksdb.NewDefaultOptions()
	opts.SetCreateIfMissing(true)
	db, err := gorocksdb.OpenDb(opts, dir)
	ensure.Nil(t, err)

	return db
}

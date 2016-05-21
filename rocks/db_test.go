package rocks

import (
	"github.com/facebookgo/ensure"
	"github.com/tecbot/gorocksdb"
	"io/ioutil"
	"log"
	"math"
	"testing"
)

func TestOpenDB(t *testing.T) {
	db := New(newRocksDB(t))
	defer db.Close()
}

func TestDBTypeOf(t *testing.T) {
	db := New(newRocksDB(t))
	defer db.Close()

	db.Set([]byte("name"), []byte("latermoon"))
	e := db.TypeOf([]byte("name"))
	ensure.True(t, e == STRING)

	e = db.TypeOf([]byte("age"))
	ensure.True(t, e == NONE)

}

func TestDBEnum(t *testing.T) {
	db := New(newRocksDB(t))
	defer db.Close()

	db.RangeEnumerate([]byte{0}, []byte{math.MaxUint8}, IterForward, func(i int, key, value []byte, quit *bool) {
		log.Println(i, string(key), string(value))
	})
}

func TestDBGetSet(t *testing.T) {
	db := New(newRocksDB(t))
	defer db.Close()

	var (
		givenKey   = []byte("name")
		givenValue = []byte("latermoon")
	)

	ensure.Nil(t, db.Set(givenKey, givenValue))

	value, err := db.Get(givenKey)
	ensure.Nil(t, err)
	ensure.DeepEqual(t, value, givenValue)
}

func TestDBRaw(t *testing.T) {
	db := New(newRocksDB(t))
	defer db.Close()

	var (
		givenKey   = []byte("name")
		givenValue = []byte("latermoon")
	)

	ensure.Nil(t, db.RawSet(givenKey, givenValue))

	value, err := db.RawGet(givenKey)
	ensure.Nil(t, err)
	ensure.DeepEqual(t, value, givenValue)

	ensure.Nil(t, db.RawDelete(givenKey))

}

func newRocksDB(t *testing.T) *gorocksdb.DB {
	dir, err := ioutil.TempDir("", "rocks")
	ensure.Nil(t, err)

	opts := gorocksdb.NewDefaultOptions()
	opts.SetCreateIfMissing(true)
	db, err := gorocksdb.OpenDb(opts, dir)
	ensure.Nil(t, err)

	return db
}

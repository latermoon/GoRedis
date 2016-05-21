package rocks

import (
	// "bytes"
	// "errors"
	// "github.com/tecbot/gorocksdb"
	"sync"
)

// list
// +key,l = ""
// l[key]0 = "a"
// l[key]1 = "b"
// l[key]2 = "c"
type ListElement struct {
	db  *DB
	key []byte
	mu  sync.RWMutex
}

package rocks

import (
	"bytes"
	"errors"
	"github.com/tecbot/gorocksdb"
	"sync"
)

// hash
// 	+key,h = ""
// 	h[key]name = "latermoon"
// 	h[key]age = "27"
// 	h[key]sex = "M"
type HashElement struct {
	db  *DB
	key []byte
	mu  sync.RWMutex
}

func NewHashElement(db *DB, key []byte) *HashElement {
	h := &HashElement{db: db, key: key}
	return h
}

func (h *HashElement) Enumerate(fn func(i int, field, value []byte, quit *bool)) {
	h.mu.RLock()
	defer h.mu.RUnlock()
	h.db.PrefixEnumerate(h.fieldPrefix(), IterForward, func(i int, key, value []byte, quit *bool) {
		fn(i, h.fieldInKey(key), value, quit)
	})
}

func (h *HashElement) Set(field, value []byte) error {
	return h.multiSet(field, value)
}

func (h *HashElement) Get(field []byte) ([]byte, error) {
	h.mu.RLock()
	defer h.mu.RUnlock()
	return h.get(field)
}

func (h *HashElement) MGet(fields ...[]byte) ([][]byte, error) {
	h.mu.RLock()
	defer h.mu.RUnlock()

	vals := make([][]byte, 0, len(fields))
	for _, field := range fields {
		val, err := h.get(field)
		if err != nil {
			return nil, err
		}
		vals = append(vals, val)
	}

	return vals, nil
}

func (h *HashElement) multiSet(fieldVals ...[]byte) error {
	h.mu.Lock()
	defer h.mu.Unlock()

	if len(fieldVals) == 0 || len(fieldVals)%2 != 0 {
		return errors.New("invalid field value pairs")
	}

	batch := gorocksdb.NewWriteBatch()
	defer batch.Destroy()
	for i := 0; i < len(fieldVals); i += 2 {
		field, value := fieldVals[i], fieldVals[i+1]
		batch.Put(h.fieldKey(field), value)
	}
	batch.Put(h.rawKey(), nil)

	return h.db.WriteBatch(batch)
}

func (h *HashElement) get(field []byte) ([]byte, error) {
	return h.db.RawGet(h.fieldKey(field))
}

func (h *HashElement) Exist(field []byte) (bool, error) {
	val, err := h.Get(field)
	if err != nil {
		return false, err
	}
	return val != nil, nil
}

func (h *HashElement) Remove(fields ...[]byte) error {
	h.mu.Lock()
	defer h.mu.Unlock()

	batch := gorocksdb.NewWriteBatch()
	defer batch.Destroy()

	dict := make(map[string]bool)
	for _, field := range fields {
		dict[string(field)] = true
		batch.Delete(h.fieldKey(field))
	}

	deleteAll := true
	h.db.PrefixEnumerate(h.fieldPrefix(), IterForward, func(i int, key, value []byte, quit *bool) {
		field := h.fieldInKey(key)
		if _, ok := dict[string(field)]; !ok { // wouldn't delete raw key
			deleteAll = false
			*quit = true
		}
	})
	if deleteAll {
		batch.Delete(h.rawKey())
	}

	return h.db.WriteBatch(batch)
}

func (h *HashElement) drop() error {
	h.mu.Lock()
	defer h.mu.Unlock()

	batch := gorocksdb.NewWriteBatch()
	defer batch.Destroy()

	h.db.PrefixEnumerate(h.fieldPrefix(), IterForward, func(i int, key, value []byte, quit *bool) {
		batch.Delete(copyBytes(key))
	})
	batch.Delete(h.rawKey())

	err := h.db.WriteBatch(batch)
	if err == nil {
		h.db = nil // make sure not invoked by others
	}
	return err
}

// +key,h
func (h *HashElement) rawKey() []byte {
	return rawKey(h.key, HASH)
}

// h[key]field
func (h *HashElement) fieldKey(field []byte) []byte {
	return bytes.Join([][]byte{h.fieldPrefix(), field}, nil)
}

// h[key]
func (h *HashElement) fieldPrefix() []byte {
	return bytes.Join([][]byte{[]byte{HASH}, SOK, h.key, EOK}, nil)
}

// split h[key]field into field
func (h *HashElement) fieldInKey(fieldKey []byte) []byte {
	right := bytes.Index(fieldKey, EOK)
	return fieldKey[right+1:]
}

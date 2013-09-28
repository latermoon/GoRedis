package storage

import (
	"sync"
)

type EntryType int

const (
	EntryTypeUnknown   = 0
	EntryTypeString    = 1
	EntryTypeHash      = 2
	EntryTypeList      = 3
	EntryTypeSet       = 4
	EntryTypeSortedSet = 5
)

// ====================Entry====================
type Entry interface {
	Value() interface{}
	Type() EntryType
	Size() int
}

type BaseEntry struct {
	value     interface{}
	entryType EntryType
}

func (b *BaseEntry) Value() interface{} {
	return b.value
}

func (b *BaseEntry) Type() EntryType {
	return b.entryType
}

func (b *BaseEntry) Size() int {
	return 0
}

// ====================StringEntry====================
type StringEntry struct {
	BaseEntry
}

func NewStringEntry(value interface{}) (e *StringEntry) {
	e = &StringEntry{}
	e.entryType = EntryTypeString
	e.value = value
	return
}

// ====================HashEntry====================
type HashEntry struct {
	BaseEntry
	Mutex sync.Mutex
}

func (h *HashEntry) Get(field string) (val interface{}) {
	val, _ = h.value.(map[string]interface{})[field]
	return
}

func (h *HashEntry) Set(field string, val interface{}) {
	h.value.(map[string]interface{})[field] = val
}

func (h *HashEntry) Map() map[string]interface{} {
	return h.value.(map[string]interface{})
}

func NewHashEntry() (e *HashEntry) {
	e = &HashEntry{}
	e.entryType = EntryTypeHash
	e.value = make(map[string]interface{})
	return
}

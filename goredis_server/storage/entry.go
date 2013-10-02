package storage

import (
	"errors"
	"github.com/ugorji/go/codec"
	"sync"
)

type EntryType int

// 数据类型
const (
	EntryTypeUnknown   = 0
	EntryTypeString    = 1
	EntryTypeHash      = 2
	EntryTypeList      = 3
	EntryTypeSet       = 4
	EntryTypeSortedSet = 5
)

var (
	mh = codec.MsgpackHandle{}
)

// ====================Entry====================
// Redis协议基本数据结构
type Entry interface {
	Type() EntryType
	Encode() (bs []byte, err error)
	Decode(bs []byte) (err error)
}

// 基本类型，简化子类代码
type BaseEntry struct {
	InnerType EntryType
}

func (b *BaseEntry) Encode() (bs []byte, err error) {
	return
}

func (b *BaseEntry) Decode(bs []byte) (err error) {
	return
}

func (b *BaseEntry) Type() EntryType {
	return b.InnerType
}

// ====================StringEntry====================
type StringEntry struct {
	BaseEntry
	value interface{}
}

func NewStringEntry(value interface{}) (e *StringEntry) {
	e = &StringEntry{}
	e.InnerType = EntryTypeString
	e.value = value
	return
}

func (s *StringEntry) Encode() (bs []byte, err error) {
	switch s.value.(type) {
	case []byte:
		bs = s.value.([]byte)
	case string:
		bs = []byte(s.value.(string))
	default:
		err = errors.New("bad string value")
	}
	return
}

func (s *StringEntry) Decode(bs []byte) (err error) {
	s.value = bs
	return
}

func (s *StringEntry) SetValue(value interface{}) {
	s.value = value
}

func (s *StringEntry) Value() (value interface{}) {
	return s.value
}

func (s *StringEntry) String() (str string) {
	switch s.value.(type) {
	case []byte:
		str = string(s.value.([]byte))
	case string:
		str = s.value.(string)
	}
	return
}

// ====================HashEntry====================
type HashEntry struct {
	BaseEntry
	table map[string]interface{}
	Mutex sync.Mutex
}

func NewHashEntry() (e *HashEntry) {
	e = &HashEntry{}
	e.InnerType = EntryTypeHash
	e.table = make(map[string]interface{})
	return
}

func (h *HashEntry) Encode() (bs []byte, err error) {
	enc := codec.NewEncoderBytes(&bs, &mh)
	err = enc.Encode(h.table)
	return
}

func (h *HashEntry) Decode(bs []byte) (err error) {
	dec := codec.NewDecoderBytes(bs, &mh)
	err = dec.Decode(&h.table)
	return
}

func (h *HashEntry) Get(field string) (val interface{}) {
	val, _ = h.table[field]
	return
}

func (h *HashEntry) Set(field string, val interface{}) {
	h.table[field] = val
}

func (h *HashEntry) Map() map[string]interface{} {
	return h.table
}

// ====================ListEntry====================
type ListEntry struct {
	BaseEntry
	sl *SafeList
}

func (l *ListEntry) List() (sl *SafeList) {
	return l.sl
}

func NewListEntry() (e *ListEntry) {
	e = &ListEntry{}
	e.InnerType = EntryTypeList
	e.sl = NewSafeList()
	return
}

// ====================SetEntry====================

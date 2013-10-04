package storage

import (
	"../libs/sortedset"
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

func (s *StringEntry) Bytes() (bs []byte) {
	switch s.value.(type) {
	case []byte:
		bs = s.value.([]byte)
	case string:
		bs = []byte(s.value.(string))
	}
	return
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

func NewListEntry() (e *ListEntry) {
	e = &ListEntry{}
	e.InnerType = EntryTypeList
	e.sl = NewSafeList()
	return
}

func (l *ListEntry) List() (sl *SafeList) {
	return l.sl
}

func (l *ListEntry) Encode() (bs []byte, err error) {
	enc := codec.NewEncoderBytes(&bs, &mh)
	arr := make([]string, 0, l.List().Len())
	for e := l.List().Front(); e != nil; e = e.Next() {
		arr = append(arr, e.Value.(string))
	}
	err = enc.Encode(arr)
	return
}

func (l *ListEntry) Decode(bs []byte) (err error) {
	dec := codec.NewDecoderBytes(bs, &mh)
	arr := make([]string, 0)
	err = dec.Decode(&arr)
	l.List().RPush(arr...)
	return
}

// ====================SetEntry====================
type SetEntry struct {
	BaseEntry
	table map[string]interface{}
	mutex sync.Mutex
}

func NewSetEntry() (s *SetEntry) {
	s = &SetEntry{}
	s.InnerType = EntryTypeSet
	s.table = make(map[string]interface{})
	return
}

// ok=true, 添加新key
// ok=false, key已存在
func (s *SetEntry) Put(key string) (ok bool) {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	_, exist := s.table[key]
	if !exist {
		s.table[key] = nil
		ok = true
	}
	return
}

func (s *SetEntry) Contains(key string) (exist bool) {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	_, exist = s.table[key]
	return
}

func (s *SetEntry) Remove(key string) (ok bool) {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	_, exist := s.table[key]
	if exist {
		delete(s.table, key)
		ok = true
	}
	return
}

func (s *SetEntry) Keys() (keys []interface{}) {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	keys = make([]interface{}, 0, len(s.table))
	for key, _ := range s.table {
		keys = append(keys, key)
	}
	return
}

func (s *SetEntry) Count() int {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	return len(s.table)
}

func (s *SetEntry) Encode() (bs []byte, err error) {
	enc := codec.NewEncoderBytes(&bs, &mh)
	err = enc.Encode(s.table)
	return
}

func (s *SetEntry) Decode(bs []byte) (err error) {
	dec := codec.NewDecoderBytes(bs, &mh)
	err = dec.Decode(&s.table)
	return
}

// ====================SortedSetEntry====================
type SortedSetEntry struct {
	BaseEntry
	zset *sortedset.SortedSet
}

func NewSortedSetEntry() (s *SortedSetEntry) {
	s = &SortedSetEntry{}
	s.InnerType = EntryTypeSortedSet
	s.zset = sortedset.NewSortedSet()
	return
}

func (s *SortedSetEntry) SortedSet() (zset *sortedset.SortedSet) {
	return s.zset
}

func (s *SortedSetEntry) Encode() (bs []byte, err error) {
	enc := codec.NewEncoderBytes(&bs, &mh)
	err = enc.Encode(s.zset.Table())
	return
}

func (s *SortedSetEntry) Decode(bs []byte) (err error) {
	dec := codec.NewDecoderBytes(bs, &mh)
	table := make(map[string]float64)
	err = dec.Decode(&table)
	// 重新构建SortedSet，成本较高，应该减少反序列化
	for member, score := range table {
		s.zset.Add(member, score)
	}
	return
}

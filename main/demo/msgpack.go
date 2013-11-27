package main

import (
	"fmt"
	"github.com/ugorji/go/codec"
)

var (
	mh = codec.MsgpackHandle{}
)

type EntryType int

type Entry interface {
	Value() interface{}
	Type() EntryType
	Size() int
}

// 基本类型，简化子类代码
type BaseEntry struct {
	InnerValue interface{} `codec:"value"`
	InnerType  EntryType   `codec:"type"`
}

func (b *BaseEntry) Value() interface{} {
	return b.InnerValue
}

func (b *BaseEntry) Type() EntryType {
	return b.InnerType
}

func (b *BaseEntry) Size() int {
	return 0
}

func NewBaseEntry() (entry Entry) {
	baseentry := &BaseEntry{}
	baseentry.InnerValue = make(map[string]interface{})
	baseentry.InnerType = 1
	return baseentry
}

func main() {
	entry := NewBaseEntry()
	m := entry.Value().(map[string]interface{})
	m["name"] = "Latermoon"
	m["age"] = 12
	var out []byte
	enc := codec.NewEncoderBytes(&out, &mh)
	e1 := enc.Encode(entry)
	fmt.Println(string(out), e1, out)

	elem2 := NewBaseEntry()
	dec := codec.NewDecoderBytes(out, &mh)
	dec.Decode(elem2)
	fmt.Println(elem2)
}

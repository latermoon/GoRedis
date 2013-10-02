package main

import (
	//"../../goredis_server/storage"
	"bytes"
	"encoding/gob"
	"fmt"
)

// 基本类型，简化子类代码
type BaseEntry struct {
	Value     interface{}
	EntryType int
}

func encodeToBytes(in interface{}) (bs []byte, err error) {
	buf := new(bytes.Buffer)
	enc := gob.NewEncoder(buf)
	err = enc.Encode(in)
	if err != nil {
		return
	}
	bs = buf.Bytes()
	return
}

func decodeFromBytes(bs []byte, out interface{}) (err error) {
	buf := new(bytes.Buffer)
	buf.Write(bs)
	dec := gob.NewDecoder(buf)
	err = dec.Decode(out)
	return
}

func main() {
	gob.GobDecoder
	entry := new(BaseEntry)
	entry.Value = map[string]interface{}{"name": "Latermoon"}
	entry.EntryType = 1
	bs, e1 := encodeToBytes(entry)
	fmt.Println(bs, e1)
}

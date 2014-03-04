package goredis_server

import (
	"encoding/binary"
	"fmt"
	"strconv"
)

func Int64ToBytes(i int64) []byte {
	var buf = make([]byte, 8)
	binary.BigEndian.PutUint64(buf, uint64(i))
	return buf
}

func BytesToInt64(buf []byte) int64 {
	return int64(binary.BigEndian.Uint64(buf))
}

func ParseInt64(b []byte) (i int64, err error) {
	i, err = strconv.ParseInt(string(b), 10, 64)
	return
}

// 将各种对象，转换为字符串形式，再转为[]byte数组
func formatByteSlice(v ...interface{}) (buf [][]byte) {
	buf = make([][]byte, 0, len(v))
	for i := 0; i < len(v); i++ {
		buf = append(buf, []byte(fmt.Sprint(v[i])))
	}
	return
}

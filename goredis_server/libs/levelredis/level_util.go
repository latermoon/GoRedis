package levelredis

import (
	"bytes"
	"encoding/binary"
	"strings"
)

// 简化字符串字节拼接
// b := []byte(strings.Join([]stirng{"1", "2", "3"}, ""))
// b := joinStringBytes("1", "2", "3")
func joinStringBytes(s ...string) []byte {
	return []byte(strings.Join(s, ""))
}

func joinBytes(b ...[]byte) []byte {
	return bytes.Join(b, nil)
}

// 复制数组
func copyBytes(src []byte) (dst []byte) {
	dst = make([]byte, len(src))
	copy(dst, src)
	return
}

// 范围判断 min <= v <= max
func between(v, min, max []byte) bool {
	return bytes.Compare(v, min) >= 0 && bytes.Compare(v, max) <= 0
}

// 使用二进制存储整形
func Int64ToBytes(i int64) []byte {
	var buf = make([]byte, 8)
	binary.BigEndian.PutUint64(buf, uint64(i))
	return buf
}

func BytesToInt64(buf []byte) int64 {
	return int64(binary.BigEndian.Uint64(buf))
}

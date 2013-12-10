package levelredis

import (
	"bytes"
	"encoding/binary"
	"strings"
)

func joinStringBytes(s ...string) []byte {
	return []byte(strings.Join(s, ""))
}

func joinBytes(b ...[]byte) []byte {
	return bytes.Join(b, nil)
}

func copyBytes(src []byte) (dst []byte) {
	dst = make([]byte, len(src))
	copy(dst, src)
	return
}

// 范围是否 min <= v <= max
func between(v, min, max []byte) bool {
	return bytes.Compare(v, min) >= 0 && bytes.Compare(v, max) <= 0
}

func Int64ToBytes(i int64) []byte {
	var buf = make([]byte, 8)
	binary.BigEndian.PutUint64(buf, uint64(i))
	return buf
}

func BytesToInt64(buf []byte) int64 {
	return int64(binary.BigEndian.Uint64(buf))
}

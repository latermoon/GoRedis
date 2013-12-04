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

func Int64ToBytes(i int64) []byte {
	var buf = make([]byte, 8)
	binary.BigEndian.PutUint64(buf, uint64(i))
	return buf
}

func BytesToInt64(buf []byte) int64 {
	return int64(binary.BigEndian.Uint64(buf))
}

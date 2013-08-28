package goredis

import (
	"bufio"
)

// 轻量的ReadBytes(delim)方法
// 使用reader.ReadByte()实现reader.ReadBytes(delim)的功能，在目标数据比较小的情况下，有较明显的优化
func lightReadBytes(reader *bufio.Reader, delim byte) (line []byte, err error) {
	err = nil
	var c byte
	// cap=4，是因为大部分场景下，redis里的数据长度不大于9999
	line = make([]byte, 0, 4)
	for {
		c, err = reader.ReadByte()
		if err != nil {
			return
		}
		// 遇到结束符
		if c == delim {
			break
		}
		line = append(line, c)
	}
	return
}

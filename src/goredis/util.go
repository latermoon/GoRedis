package goredis

import (
	"bufio"
)

// ==============================
// 各种工具方法
// ==============================

// 简化的ReadBytes(delim)方法
// reader.ReadBytes(delim)创建对象过多，使用下面方法让GoRedis多处理2k/s
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

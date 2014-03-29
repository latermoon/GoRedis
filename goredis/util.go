package goredis

import (
	"strconv"
)

// 缓存下标对应的字符串
var itoanums []string

func init() {
	itoanums = make([]string, 1000)
	for i, count := 0, len(itoanums); i < count; i++ {
		itoanums[i] = strconv.Itoa(i)
	}
}

// 经过缓存优化的itoa函数，减少strconv.Itoa的调用
func itoa(i int) string {
	if i < len(itoanums) {
		return itoanums[i]
	} else {
		return strconv.Itoa(i)
	}
}

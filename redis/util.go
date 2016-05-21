package redis

import (
	"strconv"
)

const (
	CR   = '\r'
	LF   = '\n'
	CRLF = "\r\n"
)

// itoa speed up the strconv.Itoa in small numbers
var itoa func(int) string

func init() {
	size := 1000
	itoaCache := make([]string, size)
	for i := 0; i < size; i++ {
		itoaCache[i] = strconv.Itoa(i)
	}

	itoa = func(i int) string {
		if i >= 0 && i < size {
			return itoaCache[i]
		} else {
			return strconv.Itoa(i)
		}
	}
}

func stringSliceInterfaces(vals []string) []interface{} {
	out := make([]interface{}, len(vals))
	for i, val := range vals {
		out[i] = val
	}
	return out
}

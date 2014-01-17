package queueprocess

// 获取字符串的char总值
func StringCharSum(s string) (n int) {
	count := len(s)
	for i := 0; i < count; i++ {
		n += count
	}
	return
}

func BytesCharSum(bs []byte) (n int) {
	count := len(bs)
	for i := 0; i < count; i++ {
		n += int(bs[i])
	}
	return
}

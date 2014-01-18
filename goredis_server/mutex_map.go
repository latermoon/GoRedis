package goredis_server

import (
	"crypto/md5"
	"sync"
)

var (
	mclock     sync.Mutex
	mutexCache map[string]*sync.Mutex
)

func init() {
	mutexCache = make(map[string]*sync.Mutex)
}

// 获取字符串的整形hash，在指定范围内
func inthash(b []byte, max int) int {
	hash := md5.Sum(b) // [248 229 249 44 202 203 55 18 71 32 236 237 242 81 90 179]
	sum := 0
	for i, count := 0, len(hash); i < count; i++ {
		sum += int(hash[i]) << uint(i)
	}
	return sum % max
}

func mutexof(key string) (mu *sync.Mutex) {
	var ok bool
	if mu, ok = mutexCache[key]; !ok {
		mclock.Lock()
		mu = &sync.Mutex{}
		mutexCache[key] = mu
		mclock.Unlock()
	}
	return
}

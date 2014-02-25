package goredis_server

// 本类尝试非典型写法，只暴露接口，隐藏局部变量
import (
	"crypto/md5"
	"sync"
)

// 获取字符串的整形hash，在指定范围内
var inthash func(b []byte, max int) int

// 获取一个mutex对象
var mutexof func(key string) (mu *sync.Mutex)

// implement
func init() {
	var mclock sync.Mutex
	mutexCache := make(map[string]*sync.Mutex)

	inthash = func(b []byte, max int) int {
		hash := md5.Sum(b) // [248 229 249 44 202 203 55 18 71 32 236 237 242 81 90 179]
		sum := 0
		for i, count := 0, len(hash); i < count; i++ {
			sum += int(hash[i]) << uint(i)
		}
		return sum % max
	}

	mutexof = func(key string) (mu *sync.Mutex) {
		mclock.Lock()
		var ok bool
		if mu, ok = mutexCache[key]; !ok {
			mu = &sync.Mutex{}
			mutexCache[key] = mu
		}
		mclock.Unlock()
		return
	}
}

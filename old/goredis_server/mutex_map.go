package goredis_server

// 共享锁
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

	inthash = func(b []byte, max int) int {
		hash := md5.Sum(b) // [248 229 249 44 202 203 55 18 71 32 236 237 242 81 90 179]
		sum := 0
		for i, count := 0, len(hash); i < count; i++ {
			sum += int(hash[i]) << uint(i)
		}
		return sum % max
	}

	var mclock sync.Mutex
	mucaches := make(map[string]*sync.Mutex)
	mutexof = func(key string) (mu *sync.Mutex) {
		var ok bool
		mu, ok = mucaches[key]
		if !ok {
			mclock.Lock()
			mu, ok = mucaches[key]
			if !ok {
				mu = &sync.Mutex{}
				mucaches[key] = mu
			}
			mclock.Unlock()
		}
		return
	}
}

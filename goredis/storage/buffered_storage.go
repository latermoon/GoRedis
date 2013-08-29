package storage

import ()

// BufferedStorage
// 对更新操作进行buffer
// 例子：一般mongodb的set操作5k/s，用bufferSize=100，可以变为1w/s，
// 如果bufferSize大于瞬间要操作的指令数，可以达到cache版本的10w/s
type BufferedStorage struct {
	storage   Storage
	cache     map[string]string
	cacheChan chan int // lock
	asyncChan chan int // 异步队列
}

/**
 * @param storage 要包装起来的storage，比如MongoStorage
 * @param bufferSize 队列长度，设为1时没有buffer效果，同步写入
 */
func NewBufferedStorage(storage Storage, bufferSize uint) (bs *BufferedStorage) {
	bs = &BufferedStorage{}
	bs.storage = storage
	bs.cache = make(map[string]string)
	bs.cacheChan = make(chan int, 1)
	bs.asyncChan = make(chan int, bufferSize)
	return
}

func (bs *BufferedStorage) Set(key string, value string) (err error) {
	// 写入内存
	bs.cacheChan <- 1
	bs.cache[key] = value
	<-bs.cacheChan

	// 异步处理, 容量限制
	bs.asyncChan <- 1
	go func() {
		bs.storage.Set(key, value)
		<-bs.asyncChan
	}()
	return
}

func (bs *BufferedStorage) Get(key string) (value string, err error) {
	var exists bool
	// 从内存读取
	value, exists = bs.cache[key]
	if !exists {
		// 不存在时从底层读取
		value, err = bs.storage.Get(key)
		bs.cacheChan <- 1
		bs.cache[key] = value
		<-bs.cacheChan
	}
	return
}

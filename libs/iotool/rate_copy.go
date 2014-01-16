package iotool

import (
	"container/list"
	"errors"
	// "github.com/latermoon/GoRedis/libs/stdlog"
	"io"
	"time"
)

type RateChangedCallback func(written int64, rate int)

// 限速复制
// 一般用于内网传输时，防止网卡满载，导致其它应用访问超时
func RateLimitCopy(dst io.Writer, src io.Reader, bytesInSecond int, callback RateChangedCallback) (written int64, err error) {
	if bytesInSecond < 5*1024*1024 {
		err = errors.New("bytesInSecond must larger than 5Mb")
		return
	}
	obj := &rateLimitCopy{
		dst:           dst,
		src:           src,
		bytesInSecond: bytesInSecond,
		callback:      callback,
	}
	written, err = obj.StartCopy()
	return
}

type rateLimitCopy struct {
	dst           io.Writer
	src           io.Reader
	callback      RateChangedCallback
	bytesInSecond int        // 传输速率，1秒
	written       int64      // 已拷贝字节数
	queue         *list.List // 存放最近1秒的传输字节
	sleepms       int        // 读取一个block休息的时间
	currentRate   int        //当前速率
	blocksize     int64      // 单次读取一个block的大小
	shouldStop    bool       // 停止计算器
}

func (r *rateLimitCopy) StartCopy() (int64, error) {
	// init
	r.blocksize = 500 * 1024 // 1Mb
	r.queue = list.New()
	blockcount := r.bytesInSecond / int(r.blocksize)
	r.sleepms = 1000 / blockcount

	// rate calc
	go r.rateRecordThread()
	for {
		n, err := io.CopyN(r.dst, r.src, r.blocksize)
		r.written += n
		if err == io.EOF {
			break
		}
		if err != nil {
			return r.written, err
		}
		if r.sleepms > 0 {
			time.Sleep(time.Millisecond * time.Duration(r.sleepms))
		}
	}
	r.shouldStop = true
	return r.written, nil
}

func (r *rateLimitCopy) rateRecordThread() {
	maxlen := 10
	ticker := time.NewTicker(time.Millisecond * 100)
	i := 0
	for _ = range ticker.C {
		if r.shouldStop {
			break
		}
		// 保持10个元素
		if r.queue.Len() >= 4 {
			r.queue.Remove(r.queue.Front())
		}
		r.queue.PushBack(r.written)
		// 计算速率
		if r.queue.Len() > 1 {
			first := r.queue.Front().Value.(int64)
			last := r.queue.Back().Value.(int64)
			rate := int(1.0 * (last - first) / int64(r.queue.Len()-1) * int64(maxlen))
			if rate > r.bytesInSecond {
				r.sleepms++
			} else {
				r.sleepms--
				if r.sleepms < 0 {
					r.sleepms = 0
				}
			}
			i++
			if i%maxlen == 0 {
				r.currentRate = rate
				if r.callback != nil {
					r.callback(r.written, r.currentRate)
				}
				// stdlog.Println("currentRate:", r.currentRate, "sleepms:", r.sleepms)
			}
		}
	}
}

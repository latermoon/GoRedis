package storage

import (
	"container/list"
)

// 线程安全的List，提供满足Redis List的函数
type SafeList struct {
	innerList *list.List
	lock      chan int
}

func NewSafeList() (sl *SafeList) {
	sl = &SafeList{}
	sl.innerList = list.New()
	sl.lock = make(chan int, 1)
	return
}

func (sl *SafeList) LPop() (value interface{}) {
	sl.lock <- 1
	elem := sl.innerList.Front()
	if elem != nil {
		value = elem.Value
		sl.innerList.Remove(elem)
	}
	<-sl.lock
	return
}

func (sl *SafeList) RPop() (value interface{}) {
	sl.lock <- 1
	elem := sl.innerList.Back()
	if elem != nil {
		value = elem.Value
		sl.innerList.Remove(elem)
	}
	<-sl.lock
	return
}

func (sl *SafeList) LPush(values ...string) (length int) {
	sl.lock <- 1
	for _, value := range values {
		sl.innerList.PushFront(value)
	}
	length = sl.innerList.Len()
	<-sl.lock
	return
}

func (sl *SafeList) RPush(values ...string) (length int) {
	sl.lock <- 1
	for _, value := range values {
		sl.innerList.PushBack(value)
	}
	length = sl.innerList.Len()
	<-sl.lock
	return
}

func (sl *SafeList) Len() (length int) {
	sl.lock <- 1
	length = sl.innerList.Len()
	<-sl.lock
	return
}

// 枚举实现，超大列表下性能不佳，并且lock住其它操作
func (sl *SafeList) Index(index int) (value interface{}) {
	sl.lock <- 1
	i := 0
	for e := sl.innerList.Front(); e != nil; e = e.Next() {
		if i == index {
			value = e.Value
			break
		}
		i++
	}
	<-sl.lock
	return
}

// 枚举实现，超大列表下性能不佳，并且lock住其它操作
func (sl *SafeList) Range(start int, end int) (values []interface{}) {
	sl.lock <- 1
	defer func() {
		<-sl.lock
	}()
	length := sl.innerList.Len()
	if start >= length || end < start {
		values = make([]interface{}, 0)
		return
	}
	// 确认返回数组大小
	resultsize := 0
	if length > end {
		resultsize = end - start + 1
	} else {
		resultsize = length - start
	}
	values = make([]interface{}, resultsize)
	// 填充数据
	i := 0
	for e := sl.innerList.Front(); e != nil; e = e.Next() {
		values[i] = e.Value
		i++
		if i >= resultsize {
			break
		}
	}
	return
}

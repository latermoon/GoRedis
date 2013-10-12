package storage

import (
	"container/list"
	"sync"
)

// 线程安全的List，提供满足Redis List的函数
type SafeList struct {
	innerList *list.List
	mutex     sync.Mutex
}

func NewSafeList() (sl *SafeList) {
	sl = &SafeList{}
	sl.innerList = list.New()
	return
}

func (sl *SafeList) Front() (elem *list.Element) {
	return sl.innerList.Front()
}

func (sl *SafeList) Back() (elem *list.Element) {
	return sl.innerList.Back()
}

func (sl *SafeList) LPop() (value interface{}) {
	sl.mutex.Lock()
	defer sl.mutex.Unlock()
	elem := sl.innerList.Front()
	if elem != nil {
		value = elem.Value
		sl.innerList.Remove(elem)
	}
	return
}

func (sl *SafeList) RPop() (value interface{}) {
	sl.mutex.Lock()
	defer sl.mutex.Unlock()
	elem := sl.innerList.Back()
	if elem != nil {
		value = elem.Value
		sl.innerList.Remove(elem)
	}
	return
}

func (sl *SafeList) LPush(values ...interface{}) (length int) {
	sl.mutex.Lock()
	defer sl.mutex.Unlock()
	for _, value := range values {
		sl.innerList.PushFront(value)
	}
	length = sl.innerList.Len()
	return
}

func (sl *SafeList) RPush(values ...interface{}) (length int) {
	sl.mutex.Lock()
	defer sl.mutex.Unlock()
	for _, value := range values {
		sl.innerList.PushBack(value)
	}
	length = sl.innerList.Len()
	return
}

func (sl *SafeList) Len() (length int) {
	sl.mutex.Lock()
	defer sl.mutex.Unlock()
	length = sl.innerList.Len()
	return
}

// 通过枚举实现，列表数据较大时性能不佳，并且lock住其它操作
func (sl *SafeList) Index(index int) (value interface{}) {
	sl.mutex.Lock()
	defer sl.mutex.Unlock()
	i := 0
	for e := sl.innerList.Front(); e != nil; e = e.Next() {
		if i == index {
			value = e.Value
			break
		}
		i++
	}
	return
}

// 通过枚举实现，列表数据较大时性能不佳，并且lock住其它操作
func (sl *SafeList) Range(start int, end int) (values []interface{}) {
	sl.mutex.Lock()
	defer sl.mutex.Unlock()
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
	values = make([]interface{}, 0, resultsize)
	// 填充数据
	i := 0
	for e := sl.innerList.Front(); e != nil; e = e.Next() {
		if i >= start && i <= end {
			values = append(values, e.Value)
		}
		i++
	}
	return
}

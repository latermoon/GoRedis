package util

import (
	"container/list"
	"sync"
)

type Stack struct {
	l     *list.List
	mutex *sync.Mutex
}

func NewStack() (stack *Stack) {
	stack = &Stack{}
	stack.l = list.New()
	stack.mutex = &sync.Mutex{}
	return
}

func (s *Stack) Len() int {
	return s.l.Len()
}

func (s *Stack) RPush(v interface{}) {
	s.mutex.Lock()
	s.l.PushBack(v)
	s.mutex.Unlock()
}

func (s *Stack) LPop() (v interface{}) {
	s.mutex.Lock()
	e := s.l.Front()
	if e != nil {
		v = e.Value
		s.l.Remove(e)
	}
	s.mutex.Unlock()
	return
}

func (s *Stack) LPush(v interface{}) {
	s.mutex.Lock()
	s.l.PushFront(v)
	s.mutex.Unlock()
}

func (s *Stack) RPop() (v interface{}) {
	s.mutex.Lock()
	e := s.l.Back()
	if e != nil {
		v = e.Value
		s.l.Remove(e)
	}
	s.mutex.Unlock()
	return
}

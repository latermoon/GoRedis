package goredis_server

import (
	. "../goredis"
	"container/list"
	"errors"
)

type SlaveServerManager struct {
	parent    *GoRedisServer
	slavelist *list.List
}

func NewSlaveServerManager(parent *GoRedisServer) (mgr *SlaveServerManager) {
	mgr = &SlaveServerManager{}
	mgr.parent = parent
	mgr.slavelist = list.New()
	return
}

func (s *SlaveServerManager) Add(slave *SlaveServer) (err error) {
	if len(slave.UID) > 0 && s.Slave(slave.UID) != nil {
		err = errors.New("Slave Exists, UID=" + slave.UID)
		return
	}
	s.slavelist.PushBack(slave)
	return
}

func (s *SlaveServerManager) Slave(uid string) (slave *SlaveServer) {
	for e := s.slavelist.Front(); e != nil; e = e.Next() {
		if e.Value.(*SlaveServer).UID == uid {
			slave = e.Value.(*SlaveServer)
			break
		}
	}
	return
}

func (s *SlaveServerManager) Slaves() (slaves []*SlaveServer) {
	slaves = make([]*SlaveServer, s.slavelist.Len())
	i := 0
	for e := s.slavelist.Front(); e != nil; e = e.Next() {
		slaves[i] = e.Value.(*SlaveServer)
		i++
	}
	return
}

func (s *SlaveServerManager) Remove(uid string) (slave *SlaveServer) {
	for e := s.slavelist.Front(); e != nil; e = e.Next() {
		if e.Value.(*SlaveServer).UID == uid {
			s.slavelist.Remove(e)
			return
		}
	}
	return
}

func (s *SlaveServerManager) PublishCommand(cmd *Command) {
	go func() {
		for e := s.slavelist.Front(); e != nil; e = e.Next() {
			slave := e.Value.(*SlaveServer)
			slave.SendCommand(cmd)
		}
	}()
}

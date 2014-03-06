package goredis_server

import (
	. "GoRedis/goredis"
)

// 负责传输数据到从库
// status = none/connected/disconnect
type SyncClient struct {
	session *Session
	status  string
}

func NewSyncClient(session *Session) (s *SyncClient, err error) {
	s = &SyncClient{}
	s.session = session
	return
}

func (s *SyncClient) Status() string {
	return s.status
}

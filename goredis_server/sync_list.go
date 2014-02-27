package goredis_server

import (
	. "GoRedis/goredis"
	"GoRedis/libs/levelredis"
	"bytes"
)

// 维护SYNC过程中主库的待发数据
type SyncList struct {
	seq int64 // 永远递增
	db  *levelredis.LevelRedis
}

func NewSyncList(db *levelredis.LevelRedis) (s *SyncList) {
	s = &SyncList{
		seq: -1,
		db:  db,
	}
	return
}

func (s *SyncList) seqkey(seq int64) []byte {
	return bytes.Join([][]byte{[]byte("seq:"), Int64ToBytes(seq)}, []byte(""))
}

func (s *SyncList) Push(cmd *Command) (seq int64) {
	return
}

func (s *SyncList) Seek(seq int64) (ok bool) {
	val := s.db.RawGet(s.seqkey(seq))
	ok = val != nil
	if ok {
		s.seq = seq
	}
	return
}

func (s *SyncList) Pop() (seq int64, args [][]byte) {
	val := s.db.RawGet(s.seqkey(s.seq))
	if val != nil {

	}
	return s.seq, nil
}

func (s *SyncList) Trim(seq int64) (n int64) {
	return
}

// 仅用于传输快照
type SnapList struct {
	buffer chan [][]byte
	quit   chan bool
	closed bool
}

func NewSnapList() (s *SnapList) {
	s = &SnapList{
		buffer: make(chan [][]byte),
	}
	return
}

func (s *SnapList) Push(args ...[]byte) {
	if s.closed {
		return
	}
	s.buffer <- args
}

func (s *SnapList) Pop() (args [][]byte) {
	if s.closed {
		return
	}
	select {
	case <-s.quit:
	case args = <-s.buffer:
	}
	return
}

func (s *SnapList) Close() {
	s.quit <- true
}

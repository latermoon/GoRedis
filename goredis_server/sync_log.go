package goredis_server

import (
	// . "GoRedis/goredis"
	"GoRedis/libs/levelredis"
	"GoRedis/libs/stdlog"
	"bytes"
	"sync"
)

// 同步日志，保存主库全部的更新操作
type SyncLog struct {
	db     *levelredis.LevelRedis
	seq    int64 // 永远递增
	prefix []byte
	mu     sync.RWMutex
}

func NewSyncLog(db *levelredis.LevelRedis, prefix string) (s *SyncLog) {
	s = &SyncLog{
		db:     db,
		seq:    -1,
		prefix: []byte(prefix),
	}
	s.initLastSeq()
	return
}

func (s *SyncLog) initLastSeq() {
	prefix := bytes.Join([][]byte{s.prefix, []byte(":id:")}, []byte(""))
	s.db.PrefixEnumerate(prefix, levelredis.IterBackward, func(i int, key, value []byte, quit *bool) {
		s.seq = s.splitSeqkey(key)
		*quit = true
	})
	stdlog.Println("[synclog] last seq", s.seq)
}

// sync:id:[seq] = value
func (s *SyncLog) seqkey(seq int64) []byte {
	return bytes.Join([][]byte{s.prefix, []byte(":id:"), Int64ToBytes(seq)}, []byte(""))
}

func (s *SyncLog) splitSeqkey(seqkey []byte) (seq int64) {
	b := seqkey[len(s.prefix)+len([]byte(":id:")):]
	return BytesToInt64(b)
}

func (s *SyncLog) enablekey() []byte {
	return bytes.Join([][]byte{s.prefix, []byte(":enable")}, []byte(""))
}

func (s *SyncLog) IsEnabled() bool {
	return true
}

func (s *SyncLog) Enable() {
	// s.db.RawSet(append(, value)
}

func (s *SyncLog) Disable() {

}

func (s *SyncLog) LastSeq() int64 {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.seq
}

func (s *SyncLog) Write(val []byte) (seq int64, err error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	seq = s.seq + 1
	err = s.db.RawSet(s.seqkey(seq), val)
	if err == nil {
		s.seq = seq
	}
	// stdlog.Printf("[synclog] %d, %s\n", s.seq, string(val))
	return
}

func (s *SyncLog) Read(seq int64) (val []byte, ok bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	val = s.db.RawGet(s.seqkey(seq))
	ok = val != nil
	return
}

func (s *SyncLog) Trim(seq int64) (n int64) {
	s.mu.Lock()
	defer s.mu.Unlock()
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

package goredis_server

import (
	"GoRedis/libs/levelredis"
	"GoRedis/libs/stdlog"
	"bytes"
	"errors"
	"sync"
	"time"
)

// 同步日志，保存主库全部的更新操作
type SyncLog struct {
	db      *levelredis.LevelRedis
	minseq  int64 // seq开始
	seq     int64 // seq结束, 永远递增
	maxlen  int64 // 最大长度
	enabled bool  // 是否开启了日志写入
	closed  bool  // 已关闭
	prefix  []byte
	mu      sync.RWMutex
}

func NewSyncLog(db *levelredis.LevelRedis, prefix string) (s *SyncLog) {
	s = &SyncLog{
		db:     db,
		minseq: -1,
		seq:    -1,
		maxlen: 3600 * 10000, // 3千600万
		prefix: []byte(prefix),
	}
	s.initSeq()
	go s.cleanRunloop()
	return
}

func (s *SyncLog) cleanRunloop() {
	for {
		if s.closed {
			break
		}
		if s.seq-s.minseq > s.maxlen {
			delseq := s.seq - s.maxlen
			for i := s.minseq; i < delseq; i++ {
				if s.closed {
					break
				}
				s.db.RawDel(s.seqkey(i))
				s.minseq = i + 1
			}
		}
		time.Sleep(time.Minute * 1)
	}
}

func (s *SyncLog) initSeq() {
	prefix := bytes.Join([][]byte{s.prefix, []byte(":id:")}, []byte(""))
	s.db.PrefixEnumerate(prefix, levelredis.IterForward, func(i int, key, value []byte, quit *bool) {
		s.minseq = s.splitSeqkey(key)
		*quit = true
	})
	s.db.PrefixEnumerate(prefix, levelredis.IterBackward, func(i int, key, value []byte, quit *bool) {
		s.seq = s.splitSeqkey(key)
		*quit = true
	})
	s.enabled = s.seq != -1
	if s.enabled {
		stdlog.Printf("synclog enabled, seq (%d, %d)\n", s.minseq, s.seq)
	}
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
	return s.enabled
}

func (s *SyncLog) Enable() {
	s.enabled = true
}

func (s *SyncLog) Disable() {
	s.enabled = false
}

func (s *SyncLog) MinSeq() int64 {
	return s.minseq
}

func (s *SyncLog) MaxSeq() int64 {
	return s.seq
}

func (s *SyncLog) Write(val []byte) (seq int64, err error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.closed {
		return -1, errors.New("db closed")
	}
	seq = s.seq + 1
	if s.seq == 0 { // 第一次写入同时初始化minseq
		s.minseq = s.seq
	}
	err = s.db.RawSet(s.seqkey(seq), val)
	if err == nil {
		s.seq = seq
	}
	return
}

func (s *SyncLog) Read(seq int64) (val []byte, err error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	if s.closed {
		return nil, errors.New("synclog closed")
	}
	val, err = s.db.RawGet(s.seqkey(seq))
	return
}

func (s *SyncLog) Close() {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.closed = true
	s.db.Close()
}

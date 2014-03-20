package slaveof

import (
	. "GoRedis/goredis"
	"GoRedis/libs/rdb"
	"strconv"
)

// =============================================
// 第三方rdb解释函数
// =============================================
type rdbDecoder struct {
	rdb.NopDecoder
	db       int
	i        int
	keyCount int64
	bufsize  int
	client   *SlaveClient
	// 数据缓冲
	hashEntry [][]byte
	setEntry  [][]byte
	listEntry [][]byte
	zsetEntry [][]byte
}

func newRdbDecoder(s *SlaveClient) (dec *rdbDecoder) {
	dec = &rdbDecoder{}
	dec.client = s
	dec.keyCount = 0
	dec.bufsize = 200
	return
}

func (p *rdbDecoder) StartDatabase(n int) {
	p.db = n
}

func (p *rdbDecoder) EndDatabase(n int) {
}

func (p *rdbDecoder) EndRDB() {
	p.client.rdbDecodeFinish(p.keyCount)
}

func (p *rdbDecoder) Set(key, value []byte, expiry int64) {
	cmd := NewCommand([]byte("SET"), key, value)
	p.client.rdbDecodeCommand(cmd)
	p.keyCount++
}

func (p *rdbDecoder) StartHash(key []byte, length, expiry int64) {
	p.keyCount++
}

func (p *rdbDecoder) Hset(key, field, value []byte) {
	if p.hashEntry == nil {
		p.hashEntry = make([][]byte, 0, p.bufsize+2)
		p.hashEntry = append(p.hashEntry, []byte("HMSET"), key)
	}
	p.hashEntry = append(p.hashEntry, field, value)
	if len(p.hashEntry) >= p.bufsize {
		cmd := NewCommand(p.hashEntry...)
		p.client.rdbDecodeCommand(cmd)
		p.hashEntry = nil
	}
}

// Hash
func (p *rdbDecoder) EndHash(key []byte) {
	if p.hashEntry != nil && len(p.hashEntry) > 2 {
		cmd := NewCommand(p.hashEntry...)
		p.client.rdbDecodeCommand(cmd)
		p.hashEntry = nil
	}
}

func (p *rdbDecoder) StartSet(key []byte, cardinality, expiry int64) {
	p.keyCount++
}

func (p *rdbDecoder) Sadd(key, member []byte) {
	if p.setEntry == nil {
		p.setEntry = make([][]byte, 0, p.bufsize+2)
		p.setEntry = append(p.setEntry, []byte("SADD"), key)
	}
	p.setEntry = append(p.setEntry, member)
	if len(p.setEntry) >= p.bufsize {
		cmd := NewCommand(p.setEntry...)
		p.client.rdbDecodeCommand(cmd)
		p.setEntry = nil
	}
}

// Set
func (p *rdbDecoder) EndSet(key []byte) {
	if p.setEntry != nil && len(p.setEntry) > 2 {
		cmd := NewCommand(p.setEntry...)
		p.client.rdbDecodeCommand(cmd)
		p.setEntry = nil
	}
}

func (p *rdbDecoder) StartList(key []byte, length, expiry int64) {
	p.keyCount++
	p.i = 0
}

func (p *rdbDecoder) Rpush(key, value []byte) {
	if p.listEntry == nil {
		p.listEntry = make([][]byte, 0, p.bufsize+2)
		p.listEntry = append(p.listEntry, []byte("RPUSH"), key)
	}
	p.listEntry = append(p.listEntry, value)
	if len(p.listEntry) >= p.bufsize {
		cmd := NewCommand(p.listEntry...)
		p.client.rdbDecodeCommand(cmd)
		p.listEntry = nil
	}
	p.i++
}

// List
func (p *rdbDecoder) EndList(key []byte) {
	if p.listEntry != nil && len(p.listEntry) > 2 {
		cmd := NewCommand(p.listEntry...)
		p.client.rdbDecodeCommand(cmd)
		p.listEntry = nil
	}
}

func (p *rdbDecoder) StartZSet(key []byte, cardinality, expiry int64) {
	p.keyCount++
	p.i = 0
}

func (p *rdbDecoder) Zadd(key []byte, score float64, member []byte) {
	if p.zsetEntry == nil {
		p.zsetEntry = make([][]byte, 0, p.bufsize+2)
		p.zsetEntry = append(p.zsetEntry, []byte("ZADD"), key)
	}
	p.zsetEntry = append(p.zsetEntry, []byte(strconv.FormatInt(int64(score), 10)), member)
	if len(p.zsetEntry) >= p.bufsize {
		cmd := NewCommand(p.zsetEntry...)
		p.client.rdbDecodeCommand(cmd)
		p.zsetEntry = nil
	}
	p.i++
}

// ZSet
func (p *rdbDecoder) EndZSet(key []byte) {
	if p.zsetEntry != nil && len(p.zsetEntry) > 2 {
		cmd := NewCommand(p.zsetEntry...)
		p.client.rdbDecodeCommand(cmd)
		p.zsetEntry = nil
	}
}

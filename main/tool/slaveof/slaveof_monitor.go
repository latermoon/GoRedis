package main

import (
	. "../../../goredis"
	"../../../goredis_server/libs/rdb"
	"fmt"
	"net"
	"strconv"
)

func main() {
	host := "latermoon.tj.momo.com:6379"
	conn, err := net.Dial("tcp", host)
	if err != nil {
		panic(err)
	}
	session := NewSession(conn)
	session.WriteCommand(NewCommand([]byte("SYNC")))

	readRdbFinish := false
	var c byte
	for {
		c, err = session.PeekByte()
		if err != nil {
			fmt.Printf("master gone away %s\n", session.RemoteAddr())
			break
		}
		if c == '*' {
			if cmd, e2 := session.ReadCommand(); e2 == nil {
				fmt.Println(cmd)
			} else {
				fmt.Printf("sync error %s\n", e2)
				err = e2
				break
			}
		} else if !readRdbFinish && c == '$' {
			fmt.Printf("[%s] sync rdb \n", session.RemoteAddr())
			session.ReadByte()
			var rdbsize int
			rdbsize, err = session.ReadLineInteger()
			if err != nil {
				break
			}
			fmt.Printf("[%s] rdb size %d bytes\n", session.RemoteAddr(), rdbsize)
			// read
			dec := newDecoder()
			err = rdb.Decode(session, dec)
			if err != nil {
				break
			}
			readRdbFinish = true
		} else {
			fmt.Printf("[%s] skip byte %q %s\n", session.RemoteAddr(), c, string(c))
			_, err = session.ReadByte()
			if err != nil {
				break
			}
		}
	}
}

// 第三方rdb解释函数
type rdbDecoder struct {
	db       int
	i        int
	keyCount int
	rdb.NopDecoder
	// 数据缓冲
	hashEntry [][]byte
	setEntry  [][]byte
	listEntry [][]byte
	zsetEntry [][]byte
}

func newDecoder() (dec *rdbDecoder) {
	dec = &rdbDecoder{}
	dec.keyCount = 0
	return
}

func (p *rdbDecoder) StartDatabase(n int) {
	p.db = n
}

func (p *rdbDecoder) EndDatabase(n int) {

}

func (p *rdbDecoder) EndRDB() {
	fmt.Printf("rdb end, sync %d items\n", p.keyCount)
}

// Set
func (p *rdbDecoder) Set(key, value []byte, expiry int64) {
	p.keyCount++
	fmt.Printf("[string] set %q %q\n", key, value)
}

func (p *rdbDecoder) StartHash(key []byte, length, expiry int64) {
	p.keyCount++
	p.hashEntry = make([][]byte, 0, length*2)
}

func (p *rdbDecoder) Hset(key, field, value []byte) {
	p.hashEntry = append(p.hashEntry, field)
	p.hashEntry = append(p.hashEntry, value)
	// fmt.Printf("[hash] hset %q %q %q\n", key, field, value)
}

// Hash
func (p *rdbDecoder) EndHash(key []byte) {
	fmt.Printf("[hash] %q count:%d\n", key, len(p.hashEntry)/2)
}

func (p *rdbDecoder) StartSet(key []byte, cardinality, expiry int64) {
	p.keyCount++
	p.setEntry = make([][]byte, 0, cardinality)
}

func (p *rdbDecoder) Sadd(key, member []byte) {
	p.setEntry = append(p.setEntry)
	// fmt.Printf("[set] sadd %q %q\n", key, member)
}

// Set
func (p *rdbDecoder) EndSet(key []byte) {
	fmt.Printf("[set] %q count:%d\n", key, len(p.setEntry))
}

func (p *rdbDecoder) StartList(key []byte, length, expiry int64) {
	p.keyCount++
	p.listEntry = make([][]byte, 0, length)
	p.i = 0
}

func (p *rdbDecoder) Rpush(key, value []byte) {
	p.listEntry = append(p.listEntry, value)
	p.i++
	// fmt.Printf("[list] rpush %q %q\n", key, value)
}

// List
func (p *rdbDecoder) EndList(key []byte) {
	fmt.Printf("[list] %q count:%d\n", key, len(p.listEntry))
}

func (p *rdbDecoder) StartZSet(key []byte, cardinality, expiry int64) {
	p.keyCount++
	p.zsetEntry = make([][]byte, 0, cardinality)
	p.i = 0
}

func (p *rdbDecoder) Zadd(key []byte, score float64, member []byte) {
	p.zsetEntry = append(p.zsetEntry, []byte(strconv.FormatInt(int64(score), 10)))
	p.zsetEntry = append(p.zsetEntry, member)
	p.i++
	// fmt.Printf("[zset] zadd %q %f %q\n", key, score, member)
}

// ZSet
func (p *rdbDecoder) EndZSet(key []byte) {
	fmt.Printf("[zset] %q count:%d\n", key, len(p.zsetEntry)/2)
}

package goredis_server

import (
	. "../goredis"
	"./libs/rdb"
	. "./storage"
	"fmt"
	"net"
	"strings"
)

// 从主库获取数据
// 对应 go_redis_server_sync.go
func (server *GoRedisServer) OnSLAVEOF(cmd *Command) (reply *Reply) {
	// connect master
	host := cmd.StringAtIndex(1)
	port := cmd.StringAtIndex(2)
	conn, err := net.Dial("tcp", host+":"+port)
	reply = ReplySwitch(err, StatusReply("OK"))
	if err != nil {
		return
	}

	session := NewSession(conn)
	go server.slaveOf(session)
	return
}

// -ERR wrong number of arguments for 'sync' command
func (server *GoRedisServer) slaveOf(session *Session) {
	//cmdsync := NewCommand([]byte("SYNC"), []byte("SLAVE_UID"), []byte(server.UID()))
	cmdsync := NewCommand([]byte("SYNC"))
	session.WriteCommand(cmdsync)

	// 这里代码有有点乱，可优化

	for {
		c, err := session.PeekByte()
		if err != nil {
			fmt.Println("master gone away ...")
			return
		}
		if c == '*' {
			if cmd, e2 := session.ReadCommand(); e2 == nil {
				// 这些sync回来的command，全部是更新操作，不需要返回reply
				reply := server.InvokeCommandHandler(session, cmd)
				server.syncCounters.Get("total").Incr(1)
				cmdName := strings.ToUpper(cmd.Name())
				switch cmdName {
				case "SET":
					server.syncCounters.Get("string").Incr(1)
				case "HSET":
					server.syncCounters.Get("hash").Incr(1)
				case "SADD":
					server.syncCounters.Get("set").Incr(1)
				case "RPUSH":
					server.syncCounters.Get("list").Incr(1)
				case "ZADD":
					server.syncCounters.Get("zset").Incr(1)
				case "PING":
					server.syncCounters.Get("ping").Incr(1)
				}
				fmt.Println(cmd, reply)
			} else {
				fmt.Println("sync error, ", e2)
				return
			}
		} else if c == '$' {
			fmt.Println("read rdb...")
			session.ReadByte()
			rdbsize, e3 := session.ReadLineInteger()
			if e3 != nil {
				panic(e3)
			}
			fmt.Println("rdbsize", rdbsize)
			// read
			dec := newDecoder(server)
			e2 := rdb.Decode(session, dec)
			if e2 != nil {
				panic(e2)
			}
		} else {
			fmt.Println("skip byte %q %s", c, string(c))
			session.ReadByte()
		}
	}
}

// 第三方rdb解释函数
type rdbDecoder struct {
	db int
	i  int
	rdb.NopDecoder
	server *GoRedisServer
	// 数据缓冲区
	stringEntry *StringEntry
	hashEntry   *HashEntry
	setEntry    *SetEntry
	listEntry   *ListEntry
	zsetEntry   *SortedSetEntry
}

func newDecoder(server *GoRedisServer) (dec *rdbDecoder) {
	dec = &rdbDecoder{}
	dec.server = server
	return
}

func (p *rdbDecoder) StartDatabase(n int) {
	p.db = n
}

func (p *rdbDecoder) EndDatabase(n int) {

}

func (p *rdbDecoder) EndRDB() {
	fmt.Println("End RDB")
}

func (p *rdbDecoder) Set(key, value []byte, expiry int64) {
	p.stringEntry = NewStringEntry(string(value))
	p.server.datasource.Set(string(key), p.stringEntry)
	p.server.syncCounters.Get("total").Incr(1)
	p.server.syncCounters.Get("string").Incr(1)
	fmt.Printf("db=%d [string] %q -> %q\n", p.db, key, value)
}

func (p *rdbDecoder) StartHash(key []byte, length, expiry int64) {
	p.hashEntry = NewHashEntry()
}

func (p *rdbDecoder) Hset(key, field, value []byte) {
	p.hashEntry.Set(string(field), string(value))
	//fmt.Printf("db=%d %q . %q -> %q\n", p.db, key, field, value)
}

func (p *rdbDecoder) EndHash(key []byte) {
	p.server.datasource.Set(string(key), p.hashEntry)
	p.server.syncCounters.Get("total").Incr(1)
	p.server.syncCounters.Get("hash").Incr(1)
	fmt.Printf("db=%d [hash] %q\n", p.db, key)
}

func (p *rdbDecoder) StartSet(key []byte, cardinality, expiry int64) {
	p.setEntry = NewSetEntry()
}

func (p *rdbDecoder) Sadd(key, member []byte) {
	p.setEntry.Put(string(member))
	//fmt.Printf("db=%d %q { %q }\n", p.db, key, member)
}

func (p *rdbDecoder) EndSet(key []byte) {
	p.server.datasource.Set(string(key), p.setEntry)
	p.server.syncCounters.Get("total").Incr(1)
	p.server.syncCounters.Get("set").Incr(1)
	fmt.Printf("db=%d [set] %q\n", p.db, key)
}

func (p *rdbDecoder) StartList(key []byte, length, expiry int64) {
	p.listEntry = NewListEntry()
	p.i = 0
}

func (p *rdbDecoder) Rpush(key, value []byte) {
	p.listEntry.List().RPush(string(value))
	//fmt.Printf("db=%d %q[%d] ->\n", p.db, key, p.i)
	p.i++
}

func (p *rdbDecoder) EndList(key []byte) {
	p.server.datasource.Set(string(key), p.listEntry)
	p.server.syncCounters.Get("total").Incr(1)
	p.server.syncCounters.Get("list").Incr(1)
	fmt.Printf("db=%d [list] %q\n", p.db, key)
}

func (p *rdbDecoder) StartZSet(key []byte, cardinality, expiry int64) {
	p.zsetEntry = NewSortedSetEntry()
	p.i = 0
}

func (p *rdbDecoder) Zadd(key []byte, score float64, member []byte) {
	p.zsetEntry.SortedSet().Add(string(member), score)
	//fmt.Printf("db=%d %q[%d] -> {%q, score=%g}\n", p.db, key, p.i, member, score)
	p.i++
}

func (p *rdbDecoder) EndZSet(key []byte) {
	p.server.datasource.Set(string(key), p.zsetEntry)
	p.server.syncCounters.Get("total").Incr(1)
	p.server.syncCounters.Get("zset").Incr(1)
	fmt.Printf("db=%d [zset] %q\n", p.db, key)
}

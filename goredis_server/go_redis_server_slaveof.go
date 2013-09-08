package goredis_server

import (
	"./rdb"
	"bufio"
	"fmt"
	. "github.com/latermoon/GoRedis/goredis"
	//"github.com/garyburd/redigo/redis"
	"net"
)

func (server *GoRedisServer) OnSlaveOf(cmd *Command, host string, port string) (err error) {
	var conn net.Conn
	conn, err = net.Dial("tcp", host+":"+port)
	if err != nil {
		return
	}
	fmt.Println("SlaveOf", host, port, "...")
	go server.slaveOf(conn)
	return
}

func (server *GoRedisServer) OnSYNC(cmd *Command) (err error) {

	return
}

func (server *GoRedisServer) slaveOf(conn net.Conn) {
	reader := bufio.NewReader(conn)
	conn.Write([]byte("SYNC\r\n"))

	// skip $size
	_, _ = reader.ReadBytes('\n')
	// rdb data
	e2 := rdb.Decode(reader, &decoder{})
	if e2 != nil {
		fmt.Println("Decode error", e2.Error())
		return
	}
	// find next command start
	reader.ReadBytes('*')
	// step back
	reader.UnreadByte()

	for {
		cmd, e3 := ReadCommand(reader)
		if e3 != nil {
			fmt.Println("ReadCommand error", e3.Error())
			return
		}
		fmt.Println(cmd.StringArgs()[:2])
	}
}

type decoder struct {
	db int
	i  int
	rdb.NopDecoder
}

func (p *decoder) StartDatabase(n int) {
	p.db = n
}

func (p *decoder) EndRDB() {
	fmt.Println("End RDB")
}

func (p *decoder) Set(key, value []byte, expiry int64) {
	fmt.Printf("db=%d %q -> %q\n", p.db, key, value)
}

func (p *decoder) Hset(key, field, value []byte) {
	fmt.Printf("db=%d %q . %q -> %q\n", p.db, key, field, value)
}

func (p *decoder) Sadd(key, member []byte) {
	fmt.Printf("db=%d %q { %q }\n", p.db, key, member)
}

func (p *decoder) StartList(key []byte, length, expiry int64) {
	p.i = 0
}

func (p *decoder) Rpush(key, value []byte) {
	fmt.Printf("db=%d %q[%d] ->\n", p.db, key, p.i)
	p.i++
}

func (p *decoder) StartZSet(key []byte, cardinality, expiry int64) {
	p.i = 0
}

func (p *decoder) Zadd(key []byte, score float64, member []byte) {
	fmt.Printf("db=%d %q[%d] -> {%q, score=%g}\n", p.db, key, p.i, member, score)
	p.i++
}

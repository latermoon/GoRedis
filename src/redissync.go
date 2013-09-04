package main

import (
	"../goredis/rdb"
	//"../goredis/rdb/crc64"
	"../goredis"
	"bufio"
	"fmt"
	"net"
)

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
	fmt.Printf("db=%d %q[%d] -> %q\n", p.db, key, p.i, value)
	p.i++
}

func (p *decoder) StartZSet(key []byte, cardinality, expiry int64) {
	p.i = 0
}

func (p *decoder) Zadd(key []byte, score float64, member []byte) {
	fmt.Printf("db=%d %q[%d] -> {%q, score=%g}\n", p.db, key, p.i, member, score)
	p.i++
}

func main() {
	//sync("latermoon.tj.momo.com:6388")
	sync("10.80.100.193:6348")
}

func sync(host string) {
	conn, e1 := net.Dial("tcp", host)
	if e1 != nil {
		panic(e1)
	}
	reader := bufio.NewReader(conn)

	fmt.Println("SYNC...")
	conn.Write([]byte("SYNC\r\n"))

	_, _ = reader.ReadBytes('\n')
	e2 := rdb.Decode(conn, &decoder{})
	if e2 != nil {
		panic(e2)
	}
	reader.ReadBytes('*')
	reader.UnreadByte()

	for {
		b, e4 := reader.ReadBytes('\n')
		if e4 != nil {
			panic(e4)
		}
		fmt.Print(string(b))
		fmt.Println("------------------------")
		continue
		cmd, e3 := goredis.ReadCommand(reader)
		if e3 != nil {
			panic(e3)
		}
		fmt.Println(string(cmd.Args[0]) + " " + string(cmd.Args[1]))
	}

}

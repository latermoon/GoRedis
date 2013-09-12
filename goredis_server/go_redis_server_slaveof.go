package goredis_server

import (
	. "../goredis"
	"./rdb"
	"bufio"
	"fmt"
	//"github.com/garyburd/redigo/redis"
	"net"
)

// SYNC SLAVE_UID 7cc0745b-66de-46d7-b155-321998c7c20e
func (server *GoRedisServer) OnSYNC(cmd *Command) (reply *Reply) {
	fmt.Println("[OnSYNC]", cmd.String())

	// 填充配置
	syncInfo := make(map[string]string)
	for i := 1; i < len(cmd.Args); i += 2 {
		syncInfo[cmd.StringAtIndex(i)] = cmd.StringAtIndex(i + 1)
	}
	uid, exists := syncInfo["SLAVE_UID"]
	if !exists {
		uid = ""
	}
	// 加入管理
	slave := NewSlaveServer(uid)
	server.slaveMgr.Add(slave)

	// update info
	server.ReplicationInfo.IsMaster = true

	return StatusReply("OK")
}

func (server *GoRedisServer) OnSLAVEOF(cmd *Command) (reply *Reply) {
	if server.ReplicationInfo.IsSlave {
		reply = ErrorReply("already slaveof " + server.ReplicationInfo.MasterHost + ":" + server.ReplicationInfo.MasterPort)
		return
	}

	// connect master
	host := cmd.StringAtIndex(1)
	port := cmd.StringAtIndex(2)
	conn, err := net.Dial("tcp", host+":"+port)
	reply = ReplySwitch(err, StatusReply("OK"))
	if err != nil {
		return
	}

	// update info
	server.ReplicationInfo.IsSlave = true
	server.ReplicationInfo.MasterHost = host
	server.ReplicationInfo.MasterPort = port

	go server.slaveOf(conn)
	return
}

func (server *GoRedisServer) slaveOf(conn net.Conn) {
	reader := bufio.NewReader(conn)

	syncCmd := &Command{}
	syncCmd.Args = [][]byte{[]byte("SYNC"), []byte("SLAVE_UID"), []byte(server.UID())}
	conn.Write(syncCmd.Bytes())

	/*
		// skip $size
		_, _ = reader.ReadBytes('\n')
		// rdb data
		e2 := rdb.Decode(reader, &decoder{})
		if e2 != nil {
			fmt.Println("Decode error", e2.Error())
			return
		}

	*/
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
		fmt.Println("RECV:", cmd.String())
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

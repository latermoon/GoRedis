package goredis_server

import (
	. "../goredis"
	"./rdb"

	"fmt"
	"net"
)

// SYNC SLAVE_UID 7cc0745b-66de-46d7-b155-321998c7c20e
func (server *GoRedisServer) OnSYNC(cmd *Command, session *Session) (reply *Reply) {
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
	slave := server.slaveMgr.Slave(uid)
	if slave == nil {
		slave = NewSlaveServer(uid)
		server.slaveMgr.Add(slave)
	}
	slave.SetSession(session)
	slave.Active()

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

	session := NewSession(conn)
	go server.slaveOf(session)
	return
}

// -ERR wrong number of arguments for 'sync' command
func (server *GoRedisServer) slaveOf(session *Session) {
	//cmdsync := NewCommand([]byte("SYNC"), []byte("SLAVE_UID"), []byte(server.UID()))
	cmdsync := NewCommand([]byte("SYNC"))
	session.WriteCommand(cmdsync)

	// for {
	// 	line, e1 := session.ReadByte()
	// 	if e1 != nil {
	// 		panic(e1)
	// 	}
	// 	fmt.Println(line, string(line))
	// }
	// return

	for {
		var c byte
		var err error
		if c, err = session.PeekByte(); err != nil {
			panic(err)
		}
		//fmt.Println("char:", string(c))
		if c == '*' {
			fmt.Println("read cmd...")
			if cmd, e2 := session.ReadCommand(); e2 != nil {
				panic(e2)
			} else {
				fmt.Println(cmd.Name())
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

			// if e3 := session.ReadRDB(); e3 != nil {
			// 	panic(e3)
			// } else {
			// 	fmt.Println("skip finish")
			// }
		} else {
			fmt.Println("skip byte %d %s", c, string(c))
			session.ReadByte()
		}
	}
}

type rdbDecoder struct {
	db int
	i  int
	rdb.NopDecoder
	server *GoRedisServer
}

func newDecoder(server *GoRedisServer) (dec *rdbDecoder) {
	dec = &rdbDecoder{}
	dec.server = server
	return
}

func (p *rdbDecoder) StartDatabase(n int) {
	p.db = n
}

func (p *rdbDecoder) EndRDB() {
	fmt.Println("End RDB")
}

func (p *rdbDecoder) Set(key, value []byte, expiry int64) {
	fmt.Printf("db=%d %q -> %q\n", p.db, key, value)
}

func (p *rdbDecoder) Hset(key, field, value []byte) {
	fmt.Printf("db=%d %q . %q -> %q\n", p.db, key, field, value)
}

func (p *rdbDecoder) Sadd(key, member []byte) {
	fmt.Printf("db=%d %q { %q }\n", p.db, key, member)
}

func (p *rdbDecoder) StartList(key []byte, length, expiry int64) {
	p.i = 0
}

func (p *rdbDecoder) Rpush(key, value []byte) {
	fmt.Printf("db=%d %q[%d] ->\n", p.db, key, p.i)
	p.i++
}

func (p *rdbDecoder) StartZSet(key []byte, cardinality, expiry int64) {
	p.i = 0
}

func (p *rdbDecoder) Zadd(key []byte, score float64, member []byte) {
	fmt.Printf("db=%d %q[%d] -> {%q, score=%g}\n", p.db, key, p.i, member, score)
	p.i++
}

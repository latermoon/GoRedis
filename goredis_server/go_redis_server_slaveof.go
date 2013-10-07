package goredis_server

import (
	. "../goredis"
	"./libs/rdb"
	. "./storage"
	"net"
	"strings"
)

// 从主库获取数据
// 对应 go_redis_server_sync.go
func (server *GoRedisServer) OnSLAVEOF(cmd *Command) (reply *Reply) {
	// connect master
	host := cmd.StringAtIndex(1)
	port := cmd.StringAtIndex(2)

	hostPort := host + ":" + port
	conn, err := net.Dial("tcp", hostPort)
	reply = ReplySwitch(err, StatusReply("OK"))
	if err != nil {
		return
	}

	session := NewSession(conn)
	go server.slaveOf(session)
	return
}

// 获取redis info
func (server *GoRedisServer) detectRedisInfo(session *Session, section string) (info string, err error) {
	cmdinfo := NewCommand([]byte("INFO"), []byte(section))
	session.WriteCommand(cmdinfo)
	var reply *Reply
	reply, err = session.ReadReply()
	if err == nil {
		switch reply.Value.(type) {
		case string:
			info = reply.Value.(string)
		case []byte:
			info = string(reply.Value.([]byte))
		default:
			info = reply.String()
		}
	}
	return
}

// -ERR wrong number of arguments for 'sync' command
func (server *GoRedisServer) slaveOf(session *Session) {
	// 检查是goredis还是官方redis
	info, e1 := server.detectRedisInfo(session, "server")
	if e1 != nil {
		server.stdlog.Error("[%s] slave of error %s", session.RemoteAddr(), e1)
		return
	}
	isGoRedis := strings.Index(info, "goredis_version") > 0

	var cmdsync *Command
	if isGoRedis {
		// 如果是GoRedis，需要发送自身uid，实现增量同步
		cmdsync = NewCommand([]byte("SYNC"), []byte("uid"), []byte(server.UID()))
	} else {
		cmdsync = NewCommand([]byte("SYNC"))
	}

	session.WriteCommand(cmdsync)

	// 这里代码有有点乱，可优化
	for {
		c, err := session.PeekByte()
		if err != nil {
			server.stdlog.Warn("master gone away %s", session.RemoteAddr())
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
				server.stdlog.Debug("sync recv %s %s", cmd, reply)
			} else {
				server.stdlog.Error("sync error %s", e2)
				return
			}
		} else if c == '$' {
			server.stdlog.Info("read rdb %s", session.RemoteAddr())
			session.ReadByte()
			rdbsize, e3 := session.ReadLineInteger()
			if e3 != nil {
				panic(e3)
			}
			server.stdlog.Info("rdbsize %d", rdbsize)
			// read
			dec := newDecoder(server)
			e2 := rdb.Decode(session, dec)
			if e2 != nil {
				panic(e2)
			}
		} else {
			server.stdlog.Debug("skip byte %q %s", c, string(c))
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
	p.server.stdlog.Info("rdb end")
}

func (p *rdbDecoder) Set(key, value []byte, expiry int64) {
	p.stringEntry = NewStringEntry(string(value))
	p.server.datasource.Set(string(key), p.stringEntry)
	p.server.syncCounters.Get("total").Incr(1)
	p.server.syncCounters.Get("string").Incr(1)
	p.server.stdlog.Debug("db=%d [string] %q -> %q", p.db, key, value)
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
	p.server.stdlog.Debug("db=%d [hash] %q", p.db, key)
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
	p.server.stdlog.Debug("db=%d [set] %q", p.db, key)
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
	p.server.stdlog.Debug("db=%d [list] %q", p.db, key)
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
	p.server.stdlog.Debug("db=%d [zset] %q", p.db, key)
}

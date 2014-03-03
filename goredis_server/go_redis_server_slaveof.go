package goredis_server

import (
	. "GoRedis/goredis"
	"GoRedis/libs/stdlog"
	"net"
	"strconv"
)

// 从主库获取数据
// 对应 go_redis_server_sync.go
func (server *GoRedisServer) OnSLAVEOF(session *Session, cmd *Command) (reply *Reply) {
	// connect to master
	host := cmd.StringAtIndex(1)
	port := cmd.StringAtIndex(2)
	hostPort := host + ":" + port

	conn, err := net.Dial("tcp", hostPort)
	if err != nil {
		reply = ErrorReply(err)
		stdlog.Println(err)
		return
	}
	reply = StatusReply("OK")
	// 异步处理
	masterSession := NewSession(conn)
	slaveClient, err := NewSlaveClient(server, masterSession)
	if err != nil {
		reply = ErrorReply(err)
		return
	}
	server.slavemgr.Add(slaveClient)

	isgoredis, version, err := slaveClient.MasterInfo()
	if err != nil {
		reply = ErrorReply(err)
		return
	}

	if isgoredis {
		slavelog.Printf("[M %s] slaveof %s GoRedis:%s\n", masterSession.RemoteAddr(), masterSession.RemoteAddr(), version)
	} else {
		slavelog.Printf("[M %s] slaveof %s Redis:%s\n", masterSession.RemoteAddr(), masterSession.RemoteAddr(), version)
	}

	if isgoredis {
		seq := server.masterSeq(masterSession.RemoteAddr().String())
		seqstr := strconv.FormatInt(seq, 10)
		stdlog.Printf("[M %s] uid %s, seq %d\n", masterSession.RemoteAddr(), server.UID(), seq)
		synccmd := NewCommand([]byte("SYNC"), []byte(server.UID()), []byte(seqstr))
		if err := masterSession.WriteCommand(synccmd); err != nil {
			stdlog.Printf("[M %s] sync error %s", masterSession.RemoteAddr(), err)
		}
		go server.slaveRunloop(masterSession)
	} else {
		go slaveClient.Sync(server.UID())
	}
	return
}

func (server *GoRedisServer) slaveRunloop(session *Session) {
	for {
		cmd, err := session.ReadCommand()
		if err != nil {
			stdlog.Println("[M %s] recv error %s", session.RemoteAddr(), err)
			break
		}
		server.On(session, cmd)
	}
}

func (server *GoRedisServer) masterSeq(host string) (seq int64) {
	key := "master:" + host + ":seq"
	seq = server.config.IntForKey(key, -1)
	return
}

func (server *GoRedisServer) updateMasterSeq(host string, seq int64) {
	key := "master:" + host + ":seq"
	server.config.SetInt(key, seq)
}

func (server *GoRedisServer) OnSYNC_RAW_BEG(session *Session, cmd *Command) (reply *Reply) {
	stdlog.Printf("[M %s] sync begin\n", session.RemoteAddr())
	return NOREPLY
}

// 收到来自主库的快照
func (server *GoRedisServer) OnSYNC_RAW(session *Session, cmd *Command) (reply *Reply) {
	server.OnRAW_SET(cmd)
	return NOREPLY
}

// 收取快照完成后，开始收取实时数据
func (server *GoRedisServer) OnSYNC_RAW_FIN(session *Session, cmd *Command) (reply *Reply) {
	stdlog.Printf("[M %s] SYNC_RAW_FIN\n", session.RemoteAddr())
	for {
		cmd, err := session.ReadCommand()
		if err != nil {
			stdlog.Printf("[M %s] recv err %s\n", session.RemoteAddr(), err)
			break
		}

		var seq int64
		seq, err = cmd.Int64AtIndex(1)
		if err != nil {
			stdlog.Printf("[M %s] seq err %s\n", session.RemoteAddr(), err)
			break
		}

		cmd, err = session.ReadCommand()
		if err != nil {
			stdlog.Printf("[M %s] cmd err %s\n", session.RemoteAddr(), err)
			break
		}
		// no reply
		server.On(session, cmd)
		server.updateMasterSeq(session.RemoteAddr().String(), seq)
	}
	return NOREPLY
}

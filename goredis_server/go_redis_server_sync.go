package goredis_server

import (
	. "GoRedis/goredis"
	"GoRedis/libs/levelredis"
	"GoRedis/libs/stdlog"
	"bytes"
	"strconv"
	"time"
)

// 从库的同步请求，一旦进入OnSYNC，直到主从断开才会结束当前函数
// SYNC [UID] [SEQ]
func (server *GoRedisServer) OnSYNC(session *Session, cmd *Command) (reply *Reply) {
	var seq int64 = -1
	if len(cmd.Args) >= 3 {
		var err error
		seq, err = cmd.Int64AtIndex(2)
		if err != nil {
			return ErrorReply("bad [SEQ]")
		}
	}
	stdlog.Printf("[S %s] slave %s\n", session.RemoteAddr(), cmd)

	// 第一次出现从库时才开启写日志
	if !server.synclog.IsEnabled() {
		stdlog.Println("synclog enable")
		server.synclog.Enable()
	}

	reply = NOREPLY

	if seq != -1 && (seq < server.synclog.MinSeq() || seq > server.synclog.MaxSeq()+1) {
		stdlog.Printf("[S %s] seq %d not in (%d, %d), closed\n", session.RemoteAddr(), seq, server.synclog.MinSeq(), server.synclog.MaxSeq())
		session.Close()
		return
	}

	remoteHost := session.RemoteAddr().String() // 真正的IP与端口地址
	session.SetAttribute(S_STATUS, REPL_WAIT)
	server.syncmgr.Put(remoteHost, session)
	defer func() {
		server.syncmgr.Remove(remoteHost)
	}()

	nextseq := seq // 最后要发送的seq
	// 全新同步，先发送快照
	if seq < 0 {
		var err error
		session.SetAttribute(S_STATUS, REPL_SEND_BULK)
		if nextseq, err = server.sendSnapshot(session); err != nil {
			stdlog.Printf("[S %s] snap send broken (%s)\n", remoteHost, err)
			return
		}
	}

	session.SetAttribute(S_STATUS, REPL_ONLINE)
	// 发送日志数据
	err := server.syncRunloop(session, nextseq)
	if err != nil {
		stdlog.Printf("[S %s] sync broken (%s)\n", remoteHost, err)
	}

	return
}

// nextseq，返回快照的下一条seq位置
func (server *GoRedisServer) sendSnapshot(session *Session) (nextseq int64, err error) {
	server.Suspend()                                   //挂起全部操作
	snap := server.levelRedis.DB().NewSnapshot()       // 挂起后建立快照
	defer server.levelRedis.DB().ReleaseSnapshot(snap) //
	lastseq := server.synclog.LastSeq()                // 获取当前日志序号
	server.Resume()                                    // 唤醒，如果不调用Resume，整个服务器无法继续工作

	if err = session.WriteCommand(NewCommand([]byte("SYNC_RAW_BEG"))); err != nil {
		return
	}

	// scan snapshot
	broken := false
	sendcount := 0
	sendfinish := false
	go func() {
		for {
			time.Sleep(time.Second * 1)
			if sendcount == -1 {
				break // finish
			} else if broken {
				break // cancel
			}
			if sendfinish {
				stdlog.Printf("[S %s] snap send finish, %d raw items\n", session.RemoteAddr(), sendcount)
				break
			} else {
				stdlog.Printf("[S %s] snap send %d raw items\n", session.RemoteAddr(), sendcount)
			}
		}
	}()

	// gogogo
	server.levelRedis.SnapshotEnumerate(snap, []byte{}, []byte{levelredis.MAXBYTE}, func(i int, key, value []byte, quit *bool) {
		if bytes.HasPrefix(key, []byte(PREFIX)) {
			return
		}
		cmd := NewCommand([]byte("SYNC_RAW"), key, value)
		err = session.WriteCommand(cmd)
		if err != nil {
			broken = true
			*quit = true
		}
		sendcount++
	})

	if broken {
		return -1, err
	}

	sendfinish = true

	if err = session.WriteCommand(NewCommand([]byte("SYNC_RAW_FIN"))); err != nil {
		return
	}

	nextseq = lastseq + 1
	return nextseq, nil
}

// 每发送一个SYNC_SEQ再发送一个CMD
func (server *GoRedisServer) syncRunloop(session *Session, lastseq int64) (err error) {
	if err = session.WriteCommand(NewCommand([]byte("SYNC_SEQ_BEG"))); err != nil {
		return
	}
	seq := lastseq
	deplymsec := 10
	for {
		var val []byte
		val, err = server.synclog.Read(seq)
		if err != nil {
			stdlog.Printf("[S %s] synclog read error %s\n", session.RemoteAddr(), err)
			break
		}
		if val == nil {
			time.Sleep(time.Millisecond * time.Duration(deplymsec))
			deplymsec += 10
			// 防死尸
			if deplymsec >= 100 {
				if err = session.WriteCommand(NewCommand([]byte("PING"))); err != nil {
					break
				}
			}
			continue
		} else {
			deplymsec = 10
		}

		seqstr := strconv.FormatInt(seq, 10)
		if err = session.WriteCommand(NewCommand([]byte("SYNC_SEQ"), []byte(seqstr))); err != nil {
			break
		}

		if _, err = session.Write(val); err != nil {
			break
		}
		seq++
	}
	// close
	session.Close()
	return
}

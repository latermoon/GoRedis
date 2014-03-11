package goredis_server

import (
	. "GoRedis/goredis"
	"GoRedis/libs/levelredis"
	"GoRedis/libs/stdlog"
	"bytes"
	"errors"
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

	if seq < server.synclog.MinSeq() || seq > server.synclog.MaxSeq() {
		stdlog.Printf("[S %s] seq %d not in (%d, %d), closed\n", session.RemoteAddr(), seq, server.synclog.MinSeq(), server.synclog.MaxSeq())
		session.Close()
		return
	}

	remoteHost := session.RemoteAddr().String() // 真正的IP与端口地址
	server.syncmgr.Put(remoteHost, session)
	defer server.syncmgr.Remove(remoteHost)

	lastseq := seq // 最后要发送的seq
	// 全新同步，先发送快照
	if seq < 0 {
		var err error
		if lastseq, err = server.sendSnapshot(session); err != nil {
			stdlog.Printf("[S %s] snapshot runloop broken %s\n", remoteHost, err)
			return
		}
	}

	// 发送日志数据
	err := server.syncRunloop(session, lastseq)
	if err != nil {
		stdlog.Printf("[S %s] sync runloop broken %s\n", remoteHost, err)
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

	stdlog.Printf("[S %s] snapshot, last seq %d\n", session.RemoteAddr(), lastseq)

	if err := session.WriteCommand(NewCommand([]byte("SYNC_RAW_BEG"))); err != nil {
		stdlog.Printf("[S %s] snapshot error\n", session.RemoteAddr())
		return -1, err
	}

	// scan snapshot
	broken := false
	server.levelRedis.SnapshotEnumerate(snap, []byte{}, []byte{levelredis.MAXBYTE}, func(i int, key, value []byte, quit *bool) {
		if bytes.HasPrefix(key, []byte(PREFIX)) {
			return
		}
		cmd := NewCommand([]byte("SYNC_RAW"), key, value)
		err := session.WriteCommand(cmd)
		if err != nil {
			stdlog.Printf("[S %s] snapshot error %s\n", session.RemoteAddr(), cmd)
			broken = true
			*quit = true
		}
	})

	if broken {
		return -1, errors.New("broken")
	}

	stdlog.Printf("[S %s] snapshot finish\n", session.RemoteAddr())

	if err := session.WriteCommand(NewCommand([]byte("SYNC_RAW_FIN"))); err != nil {
		stdlog.Printf("[S %s] sync error %s\n", session.RemoteAddr(), err)
		return -1, err
	}

	nextseq = lastseq + 1
	return nextseq, nil
}

// 每发送一个SYNC_SEQ再发送一个CMD
func (server *GoRedisServer) syncRunloop(session *Session, lastseq int64) (err error) {
	stdlog.Printf("[S %s] sync start seq %d\n", session.RemoteAddr(), lastseq)

	if err = session.WriteCommand(NewCommand([]byte("SYNC_SEQ_BEG"))); err != nil {
		stdlog.Printf("[S %s] sync error %s\n", session.RemoteAddr(), err)
		return
	}
	seq := lastseq
	for {
		var val []byte
		val, err = server.synclog.Read(seq)
		if err != nil {
			stdlog.Printf("[S %s] synclog read error %s\n", session.RemoteAddr(), err)
			break
		}
		if val == nil {
			time.Sleep(time.Millisecond * time.Duration(10))
			continue
		}

		seqstr := strconv.FormatInt(seq, 10)
		if err = session.WriteCommand(NewCommand([]byte("SYNC_SEQ"), []byte(seqstr))); err != nil {
			stdlog.Printf("[S %s] sync seq error %s\n", session.RemoteAddr(), err)
			break
		}

		if _, err = session.Write(val); err != nil {
			stdlog.Printf("[S %s] sync cmd error %s\n", session.RemoteAddr(), err)
			break
		}
		seq++
	}
	// close
	stdlog.Printf("[S %s] sync closed", session.RemoteAddr())
	session.Close()
	return
}

package goredis_server

import (
	. "GoRedis/goredis"
	"GoRedis/libs/levelredis"
	"GoRedis/libs/stdlog"
	"bytes"
	"strconv"
	"time"
)

// 从库的同步请求
// SYNC [UID] [SEQ]
func (server *GoRedisServer) OnSYNC(session *Session, cmd *Command) (reply *Reply) {
	var uid string
	var seq int64 = -1
	if len(cmd.Args) >= 2 {
		uid = cmd.StringAtIndex(1)
	}
	if len(cmd.Args) >= 3 {
		var err error
		seq, err = cmd.Int64AtIndex(2)
		if err != nil {
			return ErrorReply("bad [SEQ]")
		}
	}
	stdlog.Printf("[S %s] new slave uid %s, seq %d\n", session.RemoteAddr(), uid, seq)

	sc, err := NewSyncClient(session, server.directory)
	if err != nil {
		stdlog.Printf("[S %s] new slave error %s", session.RemoteAddr(), err)
		return
	}

	// 第一次出现从库时才开启写日志
	if !server.synclog.IsEnabled() {
		stdlog.Println("synclog enable")
		server.synclog.Enable()
	}

	if seq == -1 {
		go server.sendSnapshot(sc)
	} else {
		go server.syncRunloop(sc, seq)
	}

	return NOREPLY
}

func (server *GoRedisServer) sendSnapshot(sc *SyncClient) {
	server.Suspend()                                   //挂起全部操作
	snap := server.levelRedis.DB().NewSnapshot()       // 挂起更新后建立快照
	defer server.levelRedis.DB().ReleaseSnapshot(snap) //
	curseq := server.synclog.LastSeq()                 // 当前
	server.Resume()                                    // WARN 唤醒，如果不调用Resume，整个服务器无法工作

	stdlog.Printf("[S %s] new snapshot, cur seq %d\n", sc.session.RemoteAddr(), curseq)

	if err := sc.session.WriteCommand(NewCommand([]byte("SYNC_RAW_BEG"))); err != nil {
		stdlog.Printf("[S %s] snapshot error\n", sc.session.RemoteAddr())
		return
	}

	// scan snapshot
	server.levelRedis.SnapshotEnumerate(snap, []byte{}, []byte{levelredis.MAXBYTE}, func(i int, key, value []byte, quit *bool) {
		if bytes.HasPrefix(key, []byte(goredisPrefix)) {
			return
		}
		cmd := NewCommand([]byte("SYNC_RAW"), key, value)
		err := sc.session.WriteCommand(cmd)
		if err != nil {
			stdlog.Printf("[S %s] snapshot error %s\n", sc.session.RemoteAddr(), cmd)
			*quit = true
		}
	})

	if err := sc.session.WriteCommand(NewCommand([]byte("SYNC_RAW_FIN"))); err != nil {
		stdlog.Printf("[S %s] snapshot error\n", sc.session.RemoteAddr())
		return
	}

	stdlog.Printf("[S %s] snapshot finish\n", sc.session.RemoteAddr())

	// PING
	if err := sc.session.WriteCommand(NewCommand([]byte("SYNC_SEQ"), []byte(strconv.FormatInt(curseq, 10)))); err != nil {
		stdlog.Printf("[S %s] sync seq error %s\n", sc.session.RemoteAddr(), err)
		return
	}
	if err := sc.session.WriteCommand(NewCommand([]byte("PING"))); err != nil {
		stdlog.Printf("[S %s] sync ping error %s\n", sc.session.RemoteAddr(), err)
		return
	}

	curseq++
	go server.syncRunloop(sc, curseq)
}

// SEQ [SEQ]
// [CMD]
func (server *GoRedisServer) syncRunloop(sc *SyncClient, lastseq int64) {
	seq := lastseq
	for {
		val, ok := server.synclog.Read(seq)
		if !ok {
			time.Sleep(time.Millisecond * time.Duration(10))
			continue
		}

		seqstr := strconv.FormatInt(seq, 10)
		if err := sc.session.WriteCommand(NewCommand([]byte("SYNC_SEQ"), []byte(seqstr))); err != nil {
			stdlog.Printf("[S %s] sync seq error %s\n", sc.session.RemoteAddr(), err)
			break
		}

		if _, err := sc.session.Write(val); err != nil {
			stdlog.Printf("[S %s] sync cmd error %s\n", sc.session.RemoteAddr(), err)
			break
		}
		seq++
	}
}

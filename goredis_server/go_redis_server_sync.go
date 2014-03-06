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
	// var uid string
	var seq int64 = -1
	// if len(cmd.Args) >= 2 {
	// 	uid = cmd.StringAtIndex(1)
	// }
	if len(cmd.Args) >= 3 {
		var err error
		seq, err = cmd.Int64AtIndex(2)
		if err != nil {
			return ErrorReply("bad [SEQ]")
		}
	}
	stdlog.Printf("[S %s] slave %s\n", session.RemoteAddr(), cmd)

	sc, err := NewSyncClient(session)
	if err != nil {
		stdlog.Printf("[S %s] slave error %s", session.RemoteAddr(), err)
		return
	}

	// 第一次出现从库时才开启写日志
	if !server.synclog.IsEnabled() {
		stdlog.Println("synclog enable")
		server.synclog.Enable()
	}

	reply = NOREPLY

	if seq < 0 {
		go server.sendSnapshot(sc)
	} else {
		if seq >= server.synclog.MinSeq() && seq <= server.synclog.MaxSeq() {
			go server.syncRunloop(sc, seq)
		} else {
			stdlog.Printf("[S %s] seq %d not in (%d, %d), closed\n", sc.session.RemoteAddr(), seq, server.synclog.MinSeq(), server.synclog.MaxSeq())
			sc.session.Close()
		}
	}

	return
}

func (server *GoRedisServer) sendSnapshot(sc *SyncClient) {
	server.Suspend()                                   //挂起全部操作
	snap := server.levelRedis.DB().NewSnapshot()       // 挂起后建立快照
	defer server.levelRedis.DB().ReleaseSnapshot(snap) //
	curseq := server.synclog.LastSeq()                 // 获取当前日志序号
	server.Resume()                                    // 唤醒，如果不调用Resume，整个服务器无法继续工作

	stdlog.Printf("[S %s] snapshot, seq %d\n", sc.session.RemoteAddr(), curseq)

	if err := sc.session.WriteCommand(NewCommand([]byte("SYNC_RAW_BEG"))); err != nil {
		stdlog.Printf("[S %s] snapshot error\n", sc.session.RemoteAddr())
		return
	}

	// scan snapshot
	broken := false
	server.levelRedis.SnapshotEnumerate(snap, []byte{}, []byte{levelredis.MAXBYTE}, func(i int, key, value []byte, quit *bool) {
		if bytes.HasPrefix(key, []byte(PREFIX)) {
			return
		}
		cmd := NewCommand([]byte("SYNC_RAW"), key, value)
		err := sc.session.WriteCommand(cmd)
		if err != nil {
			stdlog.Printf("[S %s] snapshot error %s\n", sc.session.RemoteAddr(), cmd)
			broken = true
			*quit = true
		}
	})

	if broken {
		return
	}

	stdlog.Printf("[S %s] snapshot finish\n", sc.session.RemoteAddr())

	if err := sc.session.WriteCommand(NewCommand([]byte("SYNC_RAW_FIN"))); err != nil {
		stdlog.Printf("[S %s] sync error %s\n", sc.session.RemoteAddr(), err)
		return
	}

	curseq++
	go server.syncRunloop(sc, curseq)
}

// 每发送一个SYNC_SEQ再发送一个CMD
func (server *GoRedisServer) syncRunloop(sc *SyncClient, lastseq int64) {
	stdlog.Printf("[S %s] sync start seq %d\n", sc.session.RemoteAddr(), lastseq)

	if err := sc.session.WriteCommand(NewCommand([]byte("SYNC_SEQ_BEG"))); err != nil {
		stdlog.Printf("[S %s] sync error %s\n", sc.session.RemoteAddr(), err)
		return
	}
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

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

	if seq == -1 {
		go server.sendSnapshot(sc)
	} else {
		go server.syncRunloop(sc, seq)
	}

	return NOREPLY
}

func (server *GoRedisServer) sendSnapshot(sc *SyncClient) {
	curseq := server.synclog.LastSeq()
	stdlog.Printf("[S %s] snapshot start, cur seq %d\n", sc.session.RemoteAddr(), curseq)

	if err := sc.session.WriteCommand(NewCommand([]byte("SYNC_RAW_BEG"))); err != nil {
		stdlog.Printf("[%s] snapshot error\n", sc.session.RemoteAddr())
		return
	}

	// scan snapshot
	server.levelRedis.SnapshotEnumerate([]byte{}, []byte{levelredis.MAXBYTE}, func(i int, key, value []byte, quit *bool) {
		if bytes.HasPrefix(key, []byte(goredisPrefix)) {
			return
		}
		cmd := NewCommand([]byte("SYNC_RAW"), key, value)
		err := sc.session.WriteCommand(cmd)
		if err != nil {
			stdlog.Printf("[%s] snapshot error %s\n", sc.session.RemoteAddr(), cmd)
			*quit = true
		}
	})

	if err := sc.session.WriteCommand(NewCommand([]byte("SYNC_RAW_FIN"))); err != nil {
		stdlog.Printf("[%s] snapshot error\n", sc.session.RemoteAddr())
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

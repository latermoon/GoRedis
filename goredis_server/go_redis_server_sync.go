package goredis_server

import (
	. "GoRedis/goredis"
	"GoRedis/libs/levelredis"
	"GoRedis/libs/stdlog"
	"bytes"
	"errors"
	"fmt"
	"strconv"
	"time"
)

// S: SYNC UID [UID] PORT [PORT] SNAP [1/0] SEQ [-1/...]
func (server *GoRedisServer) OnSYNC(session *Session, cmd Command) (reply Reply) {
	stdlog.Printf("[S %s] %s\n", session.RemoteAddr(), cmd)

	args := cmd.Args()[1:]
	if len(args) < 2 || len(args)%2 != 0 {
		session.Close()
		return
	}
	for i := 0; i < len(args); i += 2 {
		// 其中有PORT
		session.SetAttribute(string(args[i]), string(args[i+1]))
	}

	if !server.synclog.IsEnabled() {
		stdlog.Println("synclog enable")
		server.synclog.Enable()
	}

	// 使用从库端口代替socket端口，标识来源
	h, _ := splitHostPort(session.RemoteAddr().String())
	remoteHost := fmt.Sprintf("%s:%s", h, session.GetAttribute("PORT"))
	session.SetAttribute(S_HOST, remoteHost)
	session.SetAttribute(S_STATUS, REPL_WAIT)

	go func() {
		server.syncmgr.Put(remoteHost, session)
		err := server.doSync(session, cmd)
		if err != nil {
			stdlog.Println("sync ", err)
		}
		session.Close()
		server.syncmgr.Remove(remoteHost)
	}()

	return NOREPLY
}

func (server *GoRedisServer) doSync(session *Session, cmd Command) (err error) {
	// snapshot
	var nextseq int64
	if session.GetAttribute("SEQ") != nil {
		nextseq, err = strconv.ParseInt(session.GetAttribute("SEQ").(string), 10, 64)
		if err != nil {
			return
		}
	}

	remoteHost := session.GetAttribute(S_HOST).(string)

	if session.GetAttribute("SNAP") != nil && session.GetAttribute("SNAP").(string) == "1" {
		session.SetAttribute(S_STATUS, REPL_SEND_BULK)
		if nextseq, err = server.sendSnapshot(session); err != nil {
			stdlog.Printf("[S %s] snap send broken (%s)\n", remoteHost, err)
			return
		}
	}

	if nextseq < 0 {
		nextseq = 0
	}
	if nextseq < server.synclog.MinSeq() || nextseq > server.synclog.MaxSeq()+1 {
		stdlog.Printf("[S %s] seq %d not in (%d, %d), closed\n", remoteHost, nextseq, server.synclog.MinSeq(), server.synclog.MaxSeq())
		return errors.New("bad seq range")
	}

	// 如果整个同步过程s

	stdlog.Println("sync online ...")
	session.SetAttribute(S_STATUS, REPL_ONLINE)
	// 发送日志数据
	err = server.syncRunloop(session, nextseq)
	if err != nil {
		stdlog.Printf("[S %s] sync broken (%s)\n", remoteHost, err)
	}
	return
}

// nextseq，返回快照的下一条seq位置
func (server *GoRedisServer) sendSnapshot(session *Session) (nextseq int64, err error) {
	server.Suspend()                     //挂起全部操作
	snap := server.levelRedis.Snapshot() // 挂起后建立只读快照
	defer snap.Close()                   // 必须释放
	lastseq := server.synclog.MaxSeq()   // 获取当前日志序号
	server.Resume()                      // 唤醒，如果不调用Resume，整个服务器无法继续工作

	if err = session.WriteCommand(NewCommand([]byte("SYNC_RAW_START"))); err != nil {
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
				stdlog.Printf("[S %s] snap send finish, %d raw items\n", session.GetAttribute(S_HOST), sendcount)
				break
			} else {
				stdlog.Printf("[S %s] snap send %d raw items\n", session.GetAttribute(S_HOST), sendcount)
			}
		}
	}()

	// gogogo
	snap.RangeEnumerate([]byte{}, []byte{levelredis.MAXBYTE}, levelredis.IterForward, func(i int, key, value []byte, quit *bool) {
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

	curseq := server.synclog.MaxSeq()
	if err = session.WriteCommand(NewCommand(formatByteSlice("SYNC_RAW_END", sendcount, lastseq, curseq)...)); err != nil {
		return
	}
	nextseq = lastseq + 1
	return nextseq, nil
}

// 每发送一个SYNC_SEQ再发送一个CMD
func (server *GoRedisServer) syncRunloop(session *Session, lastseq int64) (err error) {
	if err = session.WriteCommand(NewCommand([]byte("SYNC_SEQ_START"))); err != nil {
		return
	}
	seq := lastseq
	deplymsec := 10
	for {
		var val []byte
		val, err = server.synclog.Read(seq)
		if err != nil {
			break
		}
		if val == nil {
			time.Sleep(time.Millisecond * time.Duration(deplymsec))
			deplymsec += 10
			if deplymsec >= 1000 { // 防死尸
				if err = session.WriteCommand(NewCommand([]byte("PING"))); err != nil {
					break
				}
				deplymsec = 10
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

package goredis_server

import (
	. "GoRedis/goredis"
	"GoRedis/libs/levelredis"
	"GoRedis/libs/stdlog"
	"bufio"
	"os"
	"strings"
)

func (server *GoRedisServer) OnAOF(session *Session, cmd *Command) (reply *Reply) {
	onoff := strings.ToUpper(cmd.StringAtIndex(1))
	if onoff == "YES" {
		if server.aofwriter != nil {
			return ErrorReply("aof already inited")
		}
		go func() {
			err := server.onAOF_YES()
			if err != nil {
				stdlog.Println("aof error", err)
			}
		}()
	} else if onoff == "NO" {
		return server.onAOF_NO()
	} else {
		return ErrorReply("must be YES/NO")
	}
	return StatusReply("OK")
}

func (server *GoRedisServer) onAOF_YES() (err error) {
	if server.aofwriter == nil {
		filename := server.directory + "appendonly.aof"
		f, e := os.OpenFile(filename, os.O_WRONLY|os.O_TRUNC|os.O_CREATE, os.ModePerm)
		if e != nil {
			return e
		}
		server.aofwriter = NewAOFWriter(bufio.NewWriter(f))
		defer func() {
			f.Close()
			server.aofwriter.Close()
			server.aofwriter = nil
		}()
		stdlog.Println("aof inited")
	}
	server.Suspend()
	snap := server.levelRedis.Snapshot()
	defer snap.Close()
	// lastseq := server.synclog.MaxSeq()
	server.Resume()

	snap.KeyEnumerate([]byte(""), levelredis.IterForward, func(i int, key, keytype, value []byte, quit *bool) {
		// stdlog.Println(i, string(key), string(keytype))
		if server.aofwriter.IsClosed() {
			*quit = true
			return
		}
		switch string(keytype) {
		case "zset":
			server.aofwriter.AppendZSet(snap.GetSortedSet(string(key)))
		case "hash":
			server.aofwriter.AppendHash(snap.GetHash(string(key)))
		case "set":
			server.aofwriter.AppendSet(snap.GetSet(string(key)))
		case "list":
			server.aofwriter.AppendList(snap.GetList(string(key)))
		case "string":
			server.aofwriter.AppendString(key, value)
		case "doc":
			server.aofwriter.AppendDoc(snap.GetDoc(string(key)))
		case "none":
			stdlog.Println("bad key type", string(key), string(value))
		default:
			stdlog.Println("bad key type", string(key), string(keytype), string(value))
		}
	})
	return
}

func (server *GoRedisServer) onAOF_NO() (reply *Reply) {
	if server.aofwriter == nil {
		return ErrorReply("aof not inited")
	}
	server.aofwriter.Close()
	stdlog.Println("aof closed")
	return StatusReply("OK")
}

package goredis_server

import (
	. "GoRedis/goredis"
	"os"
	"runtime"
	"runtime/pprof"
	"strings"
)

func (server *GoRedisServer) OnGC(session *Session, cmd *Command) (reply *Reply) {
	runtime.GC()
	return StatusReply("OK")
}

// http://1234n.com/?post/wgskfs
// http://blog.golang.org/profiling-go-programs
// http://www.cnblogs.com/yjf512/archive/2012/12/27/2835331.html
func (server *GoRedisServer) OnPPROF(cmd *Command) (reply *Reply) {
	action := cmd.StringAtIndex(1)
	switch action {
	case "mem":
		f, err := os.Create(server.directory + "mem.prof")
		if err != nil {
			return ErrorReply(err)
		}
		pprof.WriteHeapProfile(f)
		f.Close()
		reply = StatusReply("OK")
	case "stop2":
		reply = StatusReply("OK")
	default:
		reply = ErrorReply("pprof [start/stop]")
	}
	return
}

// leveldb_prop "rocksdb.stats"
func (server *GoRedisServer) OnLEVELDB_PROP(cmd *Command) (reply *Reply) {
	prop := cmd.StringAtIndex(1)
	v := server.levelRedis.DB().PropertyValue(prop)
	lines := strings.Split(v, "\n")
	bulks := make([]interface{}, 0, len(lines))
	for _, line := range lines {
		if len(line) == 0 {
			continue
		}
		bulks = append(bulks, line)
	}
	reply = MultiBulksReply(bulks)
	return
}

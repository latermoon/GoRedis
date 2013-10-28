package goredis_server

import (
	. "../goredis"
	"bytes"
	"runtime"
	"strconv"
)

func (server *GoRedisServer) OnPING(cmd *Command) (reply *Reply) {
	reply = StatusReply("PONG")
	return
}

func (server *GoRedisServer) OnGC(cmd *Command) (reply *Reply) {
	server.stdlog.Info("[%s] GC() start ...")
	runtime.GC()
	server.stdlog.Info("[%s] GC() finish")
	reply = StatusReply("OK")
	return
}

func (server *GoRedisServer) OnINFO(cmd *Command) (reply *Reply) {
	// section := cmd.StringAtIndex(1)
	buf := bytes.Buffer{}
	buf.WriteString("# Server\n")
	buf.WriteString("goredis_version:0.1.1\n")
	buf.WriteString("\n")
	buf.WriteString("# Command Counter\n")
	buf.WriteString(server.cmdCounterInfo())
	buf.WriteString("\n")
	reply = BulkReply(buf.String())
	return
}

func (server *GoRedisServer) cmdCounterInfo() string {
	buf := bytes.Buffer{}
	names := server.cmdCateCounters.CounterNames()
	for _, name := range names {
		counter := server.cmdCateCounters.Get(name)
		buf.WriteString("cc_")
		buf.WriteString(name)
		buf.WriteString(":")
		buf.WriteString(strconv.Itoa(counter.Count()))
		buf.WriteString("\n")
	}
	return buf.String()
}

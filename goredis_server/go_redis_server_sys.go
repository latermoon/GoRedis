package goredis_server

import (
	. "../goredis"
	"bytes"
	"fmt"
	"runtime"
	"sort"
	"strconv"
)

func (server *GoRedisServer) OnPING(cmd *Command) (reply *Reply) {
	reply = StatusReply("PONG")
	return
}

func (server *GoRedisServer) OnGC(cmd *Command) (reply *Reply) {
	server.stdlog.Info("GC() start ...")
	runtime.GC()
	server.stdlog.Info("GC() finish")
	reply = StatusReply("OK")
	return
}

func (server *GoRedisServer) OnINFO(cmd *Command) (reply *Reply) {
	// section := cmd.StringAtIndex(1)
	buf := bytes.Buffer{}
	buf.WriteString("# Server\n")
	buf.WriteString(fmt.Sprintf("goredis_version:%s\n", VERSION))
	buf.WriteString("\n")
	buf.WriteString("# Command Counter\n")
	buf.WriteString(server.cmdCateCounterInfo())
	buf.WriteString("\n")
	buf.WriteString(server.cmdCounterInfo())
	buf.WriteString("\n")
	reply = BulkReply(buf.String())
	return
}

func (server *GoRedisServer) cmdCateCounterInfo() string {
	buf := bytes.Buffer{}
	names := server.cmdCateCounters.CounterNames()
	sort.Strings(names)
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

func (server *GoRedisServer) cmdCounterInfo() string {
	buf := bytes.Buffer{}
	names := server.cmdCounters.CounterNames()
	sort.Strings(names)
	for _, name := range names {
		counter := server.cmdCounters.Get(name)
		buf.WriteString("cmd_")
		buf.WriteString(name)
		buf.WriteString(":")
		buf.WriteString(strconv.Itoa(counter.Count()))
		buf.WriteString("\n")
	}
	return buf.String()
}

package goredis_server

import (
	. "../goredis"
	"bytes"
	"fmt"
	"runtime"
	"sort"
	"strconv"
	"strings"
)

func (server *GoRedisServer) OnINFO(cmd *Command) (reply *Reply) {
	section := strings.ToLower(cmd.StringAtIndex(1))
	switch section {
	case "memory":
		reply = BulkReply(server.memoryInfo())
	case "server":
		reply = BulkReply(server.serverInfo())
	case "command":
		reply = BulkReply(server.commandInfo())
	default:
		reply = BulkReply(server.serverInfo())
	}
	return
}

func (server *GoRedisServer) serverInfo() string {
	buf := bytes.Buffer{}
	buf.WriteString("# Server\n")
	buf.WriteString(fmt.Sprintf("goredis_version:%s\n", VERSION))
	return buf.String()
}

func (server *GoRedisServer) commandInfo() string {
	buf := bytes.Buffer{}
	buf.WriteString("# Command\n")
	names := server.cmdCounters.CounterNames()
	sort.Strings(names)
	for _, name := range names {
		counter := server.cmdCounters.Get(name)
		buf.WriteString("cmd_")
		buf.WriteString(name)
		buf.WriteString(":")
		buf.WriteString(strconv.FormatInt(counter.Count(), 10))
		buf.WriteString("\n")
	}
	return buf.String()
}

func (server *GoRedisServer) memoryInfo() string {
	buf := bytes.Buffer{}
	buf.WriteString("# Memory\n")
	var m runtime.MemStats
	runtime.ReadMemStats(&m)
	// General statistics.
	buf.WriteString(fmt.Sprint("m_Alloc:", m.Alloc, "\n"))
	buf.WriteString(fmt.Sprint("m_TotalAlloc:", m.TotalAlloc, "\n"))
	buf.WriteString(fmt.Sprint("m_Sys:", m.Sys, "\n"))
	buf.WriteString(fmt.Sprint("m_Lookups:", m.Lookups, "\n"))
	buf.WriteString(fmt.Sprint("m_Mallocs:", m.Mallocs, "\n"))
	buf.WriteString(fmt.Sprint("m_Frees:", m.Frees, "\n"))
	// Main allocation heap statistics.
	buf.WriteString(fmt.Sprint("m_HeapAlloc:", m.HeapAlloc, "\n"))
	buf.WriteString(fmt.Sprint("m_HeapSys:", m.HeapSys, "\n"))
	buf.WriteString(fmt.Sprint("m_HeapIdle:", m.HeapIdle, "\n"))
	buf.WriteString(fmt.Sprint("m_HeapInuse:", m.HeapInuse, "\n"))
	buf.WriteString(fmt.Sprint("m_HeapReleased:", m.HeapReleased, "\n"))
	buf.WriteString(fmt.Sprint("m_HeapObjects:", m.HeapObjects, "\n"))
	// Garbage collector statistics.
	buf.WriteString(fmt.Sprint("m_NextGC:", m.NextGC, "\n"))
	buf.WriteString(fmt.Sprint("m_LastGC:", m.LastGC, "\n"))
	buf.WriteString(fmt.Sprint("m_PauseTotalNs:", m.PauseTotalNs, "\n"))
	buf.WriteString(fmt.Sprint("m_PauseNs:", m.PauseNs, "\n"))
	buf.WriteString(fmt.Sprint("m_NumGC:", m.NumGC, "\n"))
	buf.WriteString(fmt.Sprint("m_EnableGC:", m.EnableGC, "\n"))
	buf.WriteString(fmt.Sprint("m_DebugGC:", m.DebugGC, "\n"))
	return buf.String()
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
		buf.WriteString(strconv.FormatInt(counter.Count(), 10))
		buf.WriteString("\n")
	}
	return buf.String()
}

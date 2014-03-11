package goredis_server

import (
	. "GoRedis/goredis"
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
	case "stats":
		reply = BulkReply(server.statsInfo())
	default:
		buf := bytes.Buffer{}
		buf.WriteString(server.serverInfo())
		buf.WriteString("\n")
		buf.WriteString(server.clientInfo())
		buf.WriteString("\n")
		buf.WriteString(server.memoryInfo())
		buf.WriteString("\n")
		buf.WriteString(server.persistenceInfo())
		buf.WriteString("\n")
		buf.WriteString(server.statsInfo())
		buf.WriteString("\n")
		buf.WriteString(server.replicationInfo())
		reply = BulkReply(buf.String())
	}
	return
}

func (server *GoRedisServer) serverInfo() string {
	buf := bytes.Buffer{}
	buf.WriteString("# Server\n")
	buf.WriteString(fmt.Sprintf("goredis_version:%s\n", server.info.Version()))
	return buf.String()
}

func (server *GoRedisServer) clientInfo() string {
	buf := bytes.Buffer{}
	buf.WriteString("# Clients\n")
	buf.WriteString(fmt.Sprintf("connected_clients:%d\n", server.info.connected_clients()))
	return buf.String()
}

func (server *GoRedisServer) commandInfo() string {
	buf := bytes.Buffer{}
	buf.WriteString("# Command\n")
	names := server.cmdCounters.Names()
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
	buf.WriteString(fmt.Sprintf("used_memory_peak:%d\n", 0))
	buf.WriteString(fmt.Sprintf("used_memory_peak_human:%d\n", 0))
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
	if false {
		buf.WriteString(fmt.Sprint("m_NextGC:", m.NextGC, "\n"))
		buf.WriteString(fmt.Sprint("m_LastGC:", m.LastGC, "\n"))
		buf.WriteString(fmt.Sprint("m_PauseTotalNs:", m.PauseTotalNs, "\n"))
		buf.WriteString(fmt.Sprint("m_PauseNs:", m.PauseNs, "\n"))
		buf.WriteString(fmt.Sprint("m_NumGC:", m.NumGC, "\n"))
		buf.WriteString(fmt.Sprint("m_EnableGC:", m.EnableGC, "\n"))
		buf.WriteString(fmt.Sprint("m_DebugGC:", m.DebugGC, "\n"))
	}
	return buf.String()
}

func (server *GoRedisServer) persistenceInfo() string {
	buf := bytes.Buffer{}
	buf.WriteString("# Persistence\n")
	return buf.String()
}

func (server *GoRedisServer) statsInfo() string {
	buf := bytes.Buffer{}
	buf.WriteString("# Stats\n")
	// buf.WriteString(fmt.Sprintf("total_connections_received:%d\n", 0))
	buf.WriteString(fmt.Sprintf("total_commands_processed:%d\n", server.info.total_commands_processed()))
	buf.WriteString(fmt.Sprintf("instantaneous_ops_per_sec:%d\n", server.info.instantaneous_ops_per_sec()))
	buf.WriteString(fmt.Sprintf("rejected_connections:%d\n", 0))
	// buf.WriteString(fmt.Sprintf("keyspace_hits:%d\n", 0))
	// buf.WriteString(fmt.Sprintf("keyspace_misses:%d\n", 0))
	return buf.String()
}

func (server *GoRedisServer) replicationInfo() string {
	buf := bytes.Buffer{}
	buf.WriteString("# Replication\n")
	buf.WriteString(fmt.Sprintf("role:%s\n", server.info.Role()))

	buf.WriteString(fmt.Sprintf("connected_slaves:%d\n", server.info.connected_slaves()))
	server.syncmgr.Enumerate(func(i int, key string, val interface{}) {
		sess := val.(*Session)
		buf.WriteString(fmt.Sprintf("slave%d:%s,%s\n", i, key, sess.GetAttribute(S_STATUS)))
	})

	buf.WriteString(fmt.Sprintf("connected_masters:%d\n", server.info.connected_masters()))
	server.slavemgr.Enumerate(func(i int, key string, val interface{}) {
		client := val.(ISlaveClient)
		sess := client.Session()
		buf.WriteString(fmt.Sprintf("master%d:%s,%s\n", i, sess.RemoteAddr(), sess.GetAttribute(S_STATUS)))
	})

	return buf.String()
}

func (server *GoRedisServer) cmdCateCounterInfo() string {
	buf := bytes.Buffer{}
	names := server.cmdCateCounters.Names()
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

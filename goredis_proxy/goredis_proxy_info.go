package goredis_proxy

import (
	. "GoRedis/goredis"
	"bytes"
	"fmt"
	"strings"
	"time"
)

func (server *GoRedisProxy) OnINFO(session *Session, cmd *Command) (reply *Reply) {
	buf := bytes.Buffer{}
	buf.WriteString("# Server\n")
	buf.WriteString(fmt.Sprintf("proxy_version:%s\n", VERSION))
	buf.WriteString(fmt.Sprintf("mode:%s\n", server.options.Mode))
	buf.WriteString("\n")

	buf.WriteString("# Master\n")
	buf.WriteString(server.remoteInfo("master", server.master))
	buf.WriteString("\n")

	buf.WriteString("# Slave0\n")
	buf.WriteString(server.remoteInfo("slave0", server.slave))
	buf.WriteString("\n")

	return BulkReply(buf.String())
}

func (server *GoRedisProxy) remoteInfo(prefix string, remote *RemoteSession) string {
	buf := bytes.Buffer{}
	buf.WriteString(fmt.Sprintf("%s:%s\n", prefix, server.remoteAddrInfo(remote)))
	upsec := time.Now().Sub(remote.Info.Uptime).Seconds()
	buf.WriteString(fmt.Sprintf("%s_uptime_in_seconds:%0.0f\n", prefix, upsec))
	buf.WriteString(fmt.Sprintf("%s_ops_per_sec:%d\n", prefix, remote.Info.Ops_per_sec))
	return buf.String()
}

func (server *GoRedisProxy) remoteAddrInfo(remote *RemoteSession) string {
	status := "OK"
	if !remote.Available() {
		status = "DISCONN"
	}
	return fmt.Sprintf("%s,%s", strings.Replace(remote.RemoteAddr(), ":", ",", 1), status)
}

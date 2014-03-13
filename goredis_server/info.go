package goredis_server

import (
	"time"
)

// 包装info的输出
type Info struct {
	server              *GoRedisServer
	ops_per_sec         int64
	last_total_commands int64
	slow_per_sec        int64
	last_slow_count     int64
	dbsizeTime          time.Time
	dbsize              int64
	uptime              time.Time
}

func NewInfo(server *GoRedisServer) *Info {
	v := &Info{
		server: server,
		uptime: time.Now(),
	}
	go v.secondTicker()
	return v
}

// 某些计数器需要计算每秒增量
func (i *Info) secondTicker() {
	ticker := time.NewTicker(time.Second * 1)
	for _ = range ticker.C {
		// ops_per_sec
		total := i.total_commands_processed()
		i.ops_per_sec, i.last_total_commands = total-i.last_total_commands, total
		// slow_per_sec
		slowCount := i.server.execCounters.Get(">30ms").Count()
		i.slow_per_sec, i.last_slow_count = slowCount-i.last_slow_count, slowCount
	}
	ticker.Stop()
}

func (i *Info) Version() string {
	return VERSION
}

func (i *Info) connected_clients() int64 {
	return i.server.counters.Get("connection").Count()
}

func (i *Info) instantaneous_ops_per_sec() int64 {
	return i.ops_per_sec
}

func (i *Info) total_commands_processed() int64 {
	return i.server.cmdCateCounters.Get("total").Count()
}

func (i *Info) Role() (role string) {
	slaveCount := i.connected_slaves()
	masterCount := i.connected_masters()
	if slaveCount > 0 && masterCount > 0 {
		role = "both"
	} else if slaveCount > 0 && masterCount == 0 {
		role = "master"
	} else if masterCount > 0 && slaveCount == 0 {
		role = "slave"
	} else {
		role = "none"
	}
	return
}

func (i *Info) connected_slaves() int {
	return i.server.syncmgr.Len()
}

func (i *Info) connected_masters() int {
	return i.server.slavemgr.Len()
}

func (i *Info) slow_ops_per_sec() int64 {
	return i.slow_per_sec
}

func (i *Info) uptime_in_seconds() int64 {
	return int64(time.Now().Sub(i.uptime).Seconds())
}

func (i *Info) db_size() int64 {
	if time.Now().Sub(i.dbsizeTime).Seconds() > 15 {
		i.dbsize = directoryTotalSize(i.server.directory + "db0")
		i.dbsizeTime = time.Now()
	}
	return i.dbsize
}

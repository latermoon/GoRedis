package goredis_server

import (
	"GoRedis/libs/levelredis"
	"GoRedis/libs/stat"
	"GoRedis/libs/stdlog"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"
)

func (server *GoRedisServer) Init() (err error) {
	// init errlog
	errlog.SetOutput(os.Stderr)

	server.initSignalNotify()

	stdlog.Println("server init ...")
	err = server.initLevelDB()
	if err != nil {
		return
	}
	err = server.initSyncLog()
	if err != nil {
		return
	}
	// __goredis:config:xxx
	server.config = NewConfig(server.levelRedis, goredisPrefix+"config:")
	// monitor
	server.initCommandMonitor(server.directory + "/cmd.log")
	server.initCommandCounterLog("string", []string{"GET", "SET", "MGET", "MSET", "INCR", "DECR", "INCRBY", "DECRBY"})
	server.initCommandCounterLog("hash", []string{"HGETALL", "HGET", "HSET", "HDEL", "HMGET", "HMSET", "HINCRBY", "HLEN"})
	server.initCommandCounterLog("set", []string{"SADD", "SCARD", "SISMEMBER", "SMEMBERS", "SREM"})
	server.initCommandCounterLog("list", []string{"LPUSH", "RPUSH", "LPOP", "RPOP", "LINDEX", "LLEN", "LRANGE", "LTRIM"})
	server.initCommandCounterLog("zset", []string{"ZADD", "ZCARD", "ZSCORE", "ZINCRBY", "ZRANGE", "ZRANGEBYSCORE", "ZRANK", "ZREM", "ZREMRANGEBYRANK", "ZREMRANGEBYSCORE", "ZREVRANGE", "ZREVRANGEBYSCORE", "ZREVRANK"})
	server.initSeqLog(server.directory + "/seq.log")
	server.initLeveldbIOLog(server.directory + "/leveldb.io.log")
	server.initLeveldbStatsLog(server.directory + "/leveldb.stats.log")
	server.initSlowlog(server.directory + "/slow.log")
	stdlog.Printf("init uid %s\n", server.UID())
	server.initConfig()
	return
}

// 处理退出事件
func (server *GoRedisServer) initSignalNotify() {
	server.sigs = make(chan os.Signal, 1)
	signal.Notify(server.sigs, syscall.SIGTERM)
	go func() {
		sig := <-server.sigs
		stdlog.Println("recv signal:", sig)
		server.levelRedis.Close()
		stdlog.Println("db closed, bye")
		os.Exit(0)
	}()
}

func (server *GoRedisServer) initConfig() {
	// slowlog-log-slower-than
	// slst := server.config.IntForKey("slowlog-log-slower-than", 100*1000)
}

func (server *GoRedisServer) initLevelDB() (err error) {
	opts := levelredis.NewOptions()
	opts.SetCache(levelredis.NewLRUCache(512 * 1024 * 1024))
	opts.SetCompression(levelredis.SnappyCompression)
	opts.SetBlockSize(32 * 1024)
	opts.SetMaxBackgroundCompactions(6)
	opts.SetWriteBufferSize(128 * 1024 * 1024)
	opts.SetMaxOpenFiles(100000)
	opts.SetCreateIfMissing(true)
	env := levelredis.NewDefaultEnv()
	env.SetBackgroundThreads(6)
	env.SetHighPriorityBackgroundThreads(2)
	opts.SetEnv(env)
	db, e1 := levelredis.Open(server.directory+"/db0", opts)
	if e1 != nil {
		return e1
	}
	server.levelRedis = levelredis.NewLevelRedis(db)
	return
}

func (server *GoRedisServer) initSyncLog() error {
	opts := levelredis.NewOptions()
	opts.SetCache(levelredis.NewLRUCache(512 * 1024 * 1024))
	opts.SetCompression(levelredis.NoCompression)
	opts.SetBlockSize(32 * 1024)
	opts.SetMaxBackgroundCompactions(2)
	opts.SetWriteBufferSize(128 * 1024 * 1024)
	opts.SetMaxOpenFiles(100000)
	opts.SetCreateIfMissing(true)
	env := levelredis.NewDefaultEnv()
	env.SetBackgroundThreads(2)
	env.SetHighPriorityBackgroundThreads(1)
	opts.SetEnv(env)
	db, e1 := levelredis.Open(server.directory+"/synclog", opts)
	if e1 != nil {
		return e1
	}
	ldb := levelredis.NewLevelRedis(db)
	server.synclog = NewSyncLog(ldb, "sync")
	return nil
}

// 命令执行监控
func (server *GoRedisServer) initCommandMonitor(path string) {
	file, err := os.OpenFile(path, os.O_RDWR|os.O_APPEND|os.O_CREATE, os.ModePerm)
	if err != nil {
		panic(err)
	}

	server.cmdMonitor = stat.New(file)
	st := server.cmdMonitor
	st.Add(stat.TextItem("time", 8, func() interface{} { return stat.TimeString() }))
	st.Add(stat.IncrItem("total", 7, func() int64 { return server.cmdCateCounters.Get("total").Count() }))
	// key, string, hash, list, ...
	for _, cate := range CommandCategoryList {
		func(name string) {
			var padding int
			if padding = len(name) + 1; padding < 7 {
				padding = 7
			}
			st.Add(stat.IncrItem(name, padding, func() int64 { return server.cmdCateCounters.Get(name).Count() }))
		}(string(cate))
	}

	st.Add(stat.TextItem("connection", 11, func() interface{} { return server.counters.Get("connection").Count() }))

	go st.Start()
}

func (server *GoRedisServer) initSeqLog(path string) {
	file, err := os.OpenFile(path, os.O_RDWR|os.O_APPEND|os.O_CREATE, os.ModePerm)
	if err != nil {
		panic(err)
	}
	st := stat.New(file)
	st.Add(stat.TextItem("time", 8, func() interface{} { return stat.TimeString() }))
	st.Add(stat.TextItem("minseq", 16, func() interface{} {
		if server.synclog.IsEnabled() {
			return server.synclog.MinSeq()
		} else {
			return "-"
		}
	}))
	st.Add(stat.TextItem("maxseq", 16, func() interface{} {
		if server.synclog.IsEnabled() {
			return server.synclog.MaxSeq()
		} else {
			return "-"
		}
	}))
	st.Add(stat.TextItem("size", 16, func() interface{} {
		if server.synclog.IsEnabled() {
			return server.synclog.MaxSeq() - server.synclog.MinSeq()
		} else {
			return "-"
		}
	}))
	go st.Start()
}

func (server *GoRedisServer) initSlowlog(path string) {
	file, err := os.OpenFile(path, os.O_RDWR|os.O_APPEND|os.O_CREATE, os.ModePerm)
	if err != nil {
		panic(err)
	}
	slowlog.SetOutput(file)
}

func (server *GoRedisServer) initLeveldbIOLog(path string) {
	// leveldb.io
	file, err := os.OpenFile(path, os.O_RDWR|os.O_APPEND|os.O_CREATE, os.ModePerm)
	if err != nil {
		panic(err)
	}
	server.leveldbStatus = stat.New(file)
	st := server.leveldbStatus
	st.Add(stat.TextItem("time", 8, func() interface{} { return stat.TimeString() }))

	// leveldb io 操作数
	ldbkeys := []string{"get", "set", "batch", "enum", "del", "lru_hit", "lru_miss"}
	for _, k := range ldbkeys {
		// pass local var to inner func()
		func(name string) {
			st.Add(stat.IncrItem(name, 10, func() int64 { return server.levelRedis.Counter(name) }))
		}(k)
	}
	go st.Start()
}

func (server *GoRedisServer) initLeveldbStatsLog(path string) {
	// leveldb.stats
	file, err := os.OpenFile(path, os.O_RDWR|os.O_APPEND|os.O_CREATE, os.ModePerm)
	if err != nil {
		panic(err)
	}
	ticker := time.NewTicker(time.Second * 5)
	go func() {
		for _ = range ticker.C {
			t := time.Now()
			tm := fmt.Sprintf("# %04d-%02d-%02d %02d:%02d:%02d\n", t.Year(), t.Month(), t.Day(), t.Hour(), t.Minute(), t.Second())
			file.WriteString(tm)
			file.WriteString(server.levelRedis.Stats())
			file.WriteString("\n")
		}
	}()
}

func (server *GoRedisServer) initCommandCounterLog(cate string, cmds []string) {
	path := fmt.Sprintf("%s/cmd.%s.log", server.directory, cate)
	file, err := os.OpenFile(path, os.O_RDWR|os.O_APPEND|os.O_CREATE, os.ModePerm)
	if err != nil {
		panic(err)
	}

	st := stat.New(file)
	st.Add(stat.TextItem("time", 8, func() interface{} { return stat.TimeString() }))
	for _, k := range cmds {
		func(cmd string) {
			var padding int
			if padding = len(cmd) + 1; padding < 8 {
				padding = 8
			}
			st.Add(stat.IncrItem(cmd, padding, func() int64 { return server.cmdCounters.Get(cmd).Count() }))
		}(k)
	}
	go st.Start()
}

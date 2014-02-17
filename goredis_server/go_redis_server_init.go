package goredis_server

import (
	"./monitor"
	"GoRedis/libs/levelredis"
	"GoRedis/libs/statlog"
	"GoRedis/libs/stdlog"
	"fmt"
	"os"
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
	// __goredis:config:xxx
	server.config = NewConfig(server.levelRedis, goredisPrefix+"config:")
	// monitor
	server.initCommandMonitor(server.directory + "/cmd.log")
	server.initCommandCounterLog("string", []string{"GET", "SET", "MGET", "MSET", "INCR", "DECR", "INCRBY", "DECRBY"})
	server.initCommandCounterLog("hash", []string{"HGETALL", "HGET", "HSET", "HDEL", "HMGET", "HMSET", "HINCRBY", "HLEN"})
	server.initCommandCounterLog("set", []string{"SADD", "SCARD", "SISMEMBER", "SMEMBERS", "SREM"})
	server.initCommandCounterLog("list", []string{"LPUSH", "RPUSH", "LPOP", "RPOP", "LINDEX", "LLEN", "LRANGE", "LTRIM"})
	server.initCommandCounterLog("zset", []string{"ZADD", "ZCARD", "ZSCORE", "ZINCRBY", "ZRANGE", "ZRANGEBYSCORE", "ZRANK", "ZREM", "ZREMRANGEBYRANK", "ZREMRANGEBYSCORE", "ZREVRANGE", "ZREVRANGEBYSCORE", "ZREVRANK"})
	server.initLeveldbIOLog(server.directory + "/leveldb.io.log")
	server.initLeveldbStatsLog(server.directory + "/leveldb.stats.log")
	server.initSlowlog(server.directory + "/slow.log")
	stdlog.Printf("init uid %s\n", server.UID())
	server.initConfig()
	// slave
	server.initSlaveSessions()
	return
}

func (server *GoRedisServer) initConfig() {
	// slowlog-log-slower-than
	// slst := server.config.IntForKey("slowlog-log-slower-than", 100*1000)
}

func (server *GoRedisServer) initLevelDB() (err error) {
	opts := levelredis.NewOptions()
	opts.SetCache(levelredis.NewLRUCache(128 * 1024 * 1024))
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

// 命令执行监控
func (server *GoRedisServer) initCommandMonitor(path string) {
	// monitor
	server.cmdMonitor = monitor.NewStatusLogger(path)
	server.cmdMonitor.Add(monitor.NewTimeFormater("time", 8))
	server.cmdMonitor.Add(monitor.NewCountFormater(server.cmdCateCounters.Get("total"), "total", 7, "ChangedCount"))
	// key, string, hash, list, ...
	for _, cate := range CommandCategoryList {
		cateName := string(cate)
		padding := len(cateName) + 1
		if padding < 7 {
			padding = 7
		}
		server.cmdMonitor.Add(monitor.NewCountFormater(server.cmdCateCounters.Get(cateName), cateName, padding, "ChangedCount"))
	}
	server.cmdMonitor.Add(monitor.NewCountFormater(server.counters.Get("connection"), "connection", 11, ""))
	go server.cmdMonitor.Start()
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
	server.leveldbStatus = statlog.NewStatLogger(file)
	server.leveldbStatus.Add(statlog.TimeItem("time"))
	// leveldb io 操作数
	ldbkeys := []string{"get", "set", "batch", "enum", "del", "lru_hit", "lru_miss"}
	opt := &statlog.Opt{Padding: 10}
	for _, k := range ldbkeys {
		// pass local var to inner func()
		func(name string) {
			server.leveldbStatus.Add(statlog.Item(name, func() interface{} {
				c := server.counters.Get("leveldb_io_" + name)
				c.SetCount(server.levelRedis.Counter(name))
				return c.ChangedCount()
			}, opt))
		}(k)
	}
	go server.leveldbStatus.Start()
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
	slog := statlog.NewStatLogger(file)
	slog.Add(statlog.TimeItem("time"))
	for _, k := range cmds {
		// pass local var to inner func()
		func(cmd string) {
			padding := len(cmd) + 1
			if padding < 8 {
				padding = 8
			}
			opt := &statlog.Opt{Padding: padding}
			slog.Add(statlog.Item(cmd, func() interface{} {
				c := server.cmdCounters.Get(cmd)
				return c.ChangedCount()
			}, opt))
		}(k)
	}
	go slog.Start()
}

// 初始化从库
func (server *GoRedisServer) initSlaveSessions() {
	// m := server.slaveIdMap()
	// server.stdlog.Info("init slaves: %s", m)
	// for uid, _ := range m {
	// 	slaveSession := NewSlaveSession(server, nil, uid)
	// 	server.slavelist.PushBack(slaveSession)
	// 	slaveSession.ContinueAof()
	// }
}

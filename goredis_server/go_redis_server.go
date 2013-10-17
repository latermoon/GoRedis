package goredis_server

import (
	. "../goredis"
	"./libs/golog"
	"./libs/leveltool"
	"./libs/uuid"
	"./monitor"
	. "./storage"
	"container/list"
	"errors"
	"strings"
	"sync"
)

var (
	WrongKindError = errors.New("Wrong kind opration")
	WrongKindReply = ErrorReply(WrongKindError)
)

// 数据类型描述
var entryTypeDesc = map[EntryType]string{
	EntryTypeUnknown:   "unknown",
	EntryTypeString:    "string",
	EntryTypeHash:      "hash",
	EntryTypeList:      "list",
	EntryTypeSet:       "set",
	EntryTypeSortedSet: "zset"}

var goredisPrefix string = "__goredis:"

// GoRedisServer
type GoRedisServer struct {
	CommandHandler
	RedisServer
	// 数据源
	directory  string
	datasource DataSource
	// counters
	cmdCounters  *monitor.Counters
	syncCounters *monitor.Counters
	// logger
	statusLogger *monitor.StatusLogger
	syncMonitor  *monitor.StatusLogger
	stdlog       *golog.Logger
	// 从库
	uid              string // 实例id
	slavelist        *list.List
	needSyncCmdTable map[string]bool // 需要同步的指令
	// locks
	stringMutex sync.Mutex
	// aof
	aoftable      map[string]*leveltool.LevelList
	aoftableMutex sync.Mutex
}

/*
	server := NewGoRedisServer(directory)
	server.Init()
	server.Listen(host)
*/
func NewGoRedisServer(directory string) (server *GoRedisServer) {
	server = &GoRedisServer{}
	// set as itself
	server.SetHandler(server)
	// default datasource
	server.directory = directory
	server.needSyncCmdTable = make(map[string]bool)
	server.slavelist = list.New()
	server.aoftable = make(map[string]*leveltool.LevelList) // aof
	for _, cmd := range needSyncCmds {
		server.needSyncCmdTable[strings.ToUpper(cmd)] = true
	}
	// counter
	server.cmdCounters = monitor.NewCounters()
	server.syncCounters = monitor.NewCounters()
	return
}

func (server *GoRedisServer) Init() {
	server.initLogger()
	server.stdlog.Info("========================================")
	server.stdlog.Info("server init ...")
	// leveldb
	var e1 error
	server.datasource, e1 = NewLevelDBDataSource(server.directory + "/db0")
	if e1 != nil {
		panic(e1)
	}
	//server.datasource = NewBufferDataSource(server.datasource)
	// monitor
	server.initCommandMonitor(server.directory + "/cmd.log")
	server.initSyncMonitor(server.directory + "/sync.log")
	// slave
	server.stdlog.Info("init uid %s", server.UID())
	server.initSlaveSessions()
}

func (server *GoRedisServer) Listen(host string) {
	stdlog.Info("listen %s", host)
	server.RedisServer.Listen(host)
}

func (server *GoRedisServer) UID() (uid string) {
	if len(server.uid) == 0 {
		uidkey := "__goredis:uid"
		entry := server.datasource.Get(uidkey)
		if entry == nil {
			server.uid = uuid.UUID(8)
			entry = NewStringEntry(server.uid)
			server.datasource.Set(uidkey, entry)
		} else {
			server.uid = entry.(*StringEntry).String()
		}
	}
	return server.uid
}

func (server *GoRedisServer) StdLog() *golog.Logger {
	return server.stdlog
}

func (server *GoRedisServer) initLogger() {
	out := golog.NewConsoleAndFileWriter(server.directory + "/stdout.log")
	server.stdlog = golog.New(out, golog.DEBUG)
	// package内的全局变量，方便调用
	stdlog = server.stdlog
}

// 初始化从库
func (server *GoRedisServer) initSlaveSessions() {
	slavesEntry := server.slavesEntry()
	server.stdlog.Info("init slaves: %s", slavesEntry.Keys())
	for _, slaveuid := range slavesEntry.Keys() {
		uid := slaveuid.(string)
		slaveSession := NewSlaveSession(server, nil, uid)
		server.slavelist.PushBack(slaveSession)
		slaveSession.ContinueAof()
	}
}

// 命令执行监控
func (server *GoRedisServer) initCommandMonitor(path string) {
	// monitor
	server.statusLogger = monitor.NewStatusLogger(path)
	server.statusLogger.Add(monitor.NewTimeFormater("Time", 8))
	cmds := []string{"TOTAL", "GET", "SET", "HSET", "HGET", "HGETALL", "INCR", "DEL", "ZADD"}
	for _, cmd := range cmds {
		padding := len(cmd) + 1
		if padding < 7 {
			padding = 7
		}
		server.statusLogger.Add(monitor.NewCountFormater(server.cmdCounters.Get(cmd), cmd, padding, "ChangedCount"))
	}
	go server.statusLogger.Start()
}

// 从库同步监控
func (server *GoRedisServer) initSyncMonitor(path string) {
	server.syncMonitor = monitor.NewStatusLogger(path)
	server.syncMonitor.Add(monitor.NewTimeFormater("Time", 8))
	cmds := []string{"total", "string", "hash", "set", "list", "zset", "ping"}
	for _, cmd := range cmds {
		server.syncMonitor.Add(monitor.NewCountFormater(server.syncCounters.Get(cmd), cmd, 8, "ChangedCount"))
	}
	// buffer用于显示同步过程中的taskqueue buffer长度
	server.syncMonitor.Add(monitor.NewCountFormater(server.syncCounters.Get("buffer"), "buffer", 9, "Count"))
	go server.syncMonitor.Start()
}

// for CommandHandler
func (server *GoRedisServer) On(session *Session, cmd *Command) {
	go func() {
		cmdName := strings.ToUpper(cmd.Name())
		server.cmdCounters.Get(cmdName).Incr(1)
		server.cmdCounters.Get("TOTAL").Incr(1)

		// 同步到从库
		if _, ok := server.needSyncCmdTable[cmdName]; ok {
			for e := server.slavelist.Front(); e != nil; e = e.Next() {
				e.Value.(*SlaveSession).AsyncSendCommand(cmd)
			}
		}
	}()
}

func (server *GoRedisServer) OnUndefined(session *Session, cmd *Command) (reply *Reply) {
	return ErrorReply("Not Supported: " + cmd.String())
}

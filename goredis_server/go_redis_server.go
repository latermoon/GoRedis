package goredis_server

import (
	. "../goredis"
	"./libs/golog"
	"./libs/levelredis"
	"./libs/uuid"
	"./monitor"
	"container/list"
	"errors"
	"github.com/latermoon/levigo"
	"strings"
	"sync"
)

// 版本号，每次更新都需要升级一下
const VERSION = "1.0.6"

var (
	WrongKindError = errors.New("Wrong kind opration")
	WrongKindReply = ErrorReply(WrongKindError)
)

var goredisPrefix string = "__goredis:"

// GoRedisServer
type GoRedisServer struct {
	CommandHandler
	RedisServer
	// 数据源
	directory  string
	levelRedis *levelredis.LevelRedis
	config     *levelredis.LevelConfig
	// counters
	cmdCounters     *monitor.Counters
	cmdCateCounters *monitor.Counters // 指令集统计
	// logger
	cmdMonitor *monitor.StatusLogger
	stdlog     *golog.Logger
	// 从库
	uid              string // 实例id
	slavelist        *list.List
	needSyncCmdTable map[string]bool // 需要同步的指令
	// locks
	stringMutex sync.Mutex
	// monitor
	monitorlist  *list.List
	monitorMutex sync.Mutex
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
	server.monitorlist = list.New()
	for _, cmd := range needSyncCmds {
		server.needSyncCmdTable[strings.ToUpper(cmd)] = true
	}
	// counter
	server.cmdCounters = monitor.NewCounters()
	server.cmdCateCounters = monitor.NewCounters()
	return
}

func (server *GoRedisServer) Init() (err error) {
	server.initLogger()
	server.stdlog.Info("========================================")
	server.stdlog.Info("server init ...")
	// leveldb
	// options := opt.Options{
	// 	MaxOpenFiles: 100000,
	// 	BlockCache:   cache.NewLRUCache(32 * opt.MiB),
	// 	BlockSize:    32 * opt.KiB,
	// 	WriteBuffer:  120 << 20,
	// }
	// server.db, err = leveldb.OpenFile(server.directory+"/db0", &options)
	// if err != nil {
	// 	panic(err)
	// }
	// server.levelRedis = leveltool.NewLevelRedis(server.db)
	err = server.initLevelDB()
	if err != nil {
		return
	}
	server.config = levelredis.NewLevelConfig(server.levelRedis, goredisPrefix+"config:")
	// monitor
	server.initCommandMonitor(server.directory + "/cmd.log")
	// slave
	server.stdlog.Info("init uid %s", server.UID())
	server.initSlaveSessions()
	return
}

func (server *GoRedisServer) initLevelDB() (err error) {
	opts := levigo.NewOptions()
	opts.SetCache(levigo.NewLRUCache(32 * 1024 * 1024))
	opts.SetCompression(levigo.SnappyCompression)
	opts.SetBlockSize(32 * 1024)
	opts.SetWriteBufferSize(128 * 1024 * 1024)
	opts.SetCreateIfMissing(true)
	db, e1 := levigo.Open(server.directory+"/db0", opts)
	if e1 != nil {
		return e1
	}
	server.levelRedis = levelredis.NewLevelRedis(db)
	return
}

func (server *GoRedisServer) Listen(host string) {
	stdlog.Info("listen %s", host)
	server.RedisServer.Listen(host)
}

func (server *GoRedisServer) UID() (uid string) {
	if len(server.uid) == 0 {
		uidkey := "uid"
		server.uid = server.config.GetString(uidkey)
		if len(server.uid) == 0 {
			server.uid = uuid.UUID(8)
			server.config.SetString(uidkey, server.uid)
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
	// m := server.slaveIdMap()
	// server.stdlog.Info("init slaves: %s", m)
	// for uid, _ := range m {
	// 	slaveSession := NewSlaveSession(server, nil, uid)
	// 	server.slavelist.PushBack(slaveSession)
	// 	slaveSession.ContinueAof()
	// }
}

// 命令执行监控
func (server *GoRedisServer) initCommandMonitor(path string) {
	// monitor
	server.cmdMonitor = monitor.NewStatusLogger(path)
	server.cmdMonitor.Add(monitor.NewTimeFormater("time", 8))
	server.cmdMonitor.Add(monitor.NewCountFormater(server.cmdCounters.Get("total"), "total", 7, "ChangedCount"))
	// key, string, hash, list, ...
	for _, cate := range CommandCategoryList {
		cateName := string(cate)
		padding := len(cateName) + 1
		if padding < 7 {
			padding = 7
		}
		server.cmdMonitor.Add(monitor.NewCountFormater(server.cmdCateCounters.Get(cateName), cateName, padding, "ChangedCount"))
	}
	go server.cmdMonitor.Start()
}

// for CommandHandler
func (server *GoRedisServer) On(session *Session, cmd *Command) {
	go func() {
		cmdName := strings.ToUpper(cmd.Name())
		server.cmdCounters.Get(cmdName).Incr(1)
		cate := GetCommandCategory(cmdName)
		server.cmdCateCounters.Get(string(cate)).Incr(1)
		server.cmdCounters.Get("total").Incr(1)

		// 同步到从库
		if _, ok := server.needSyncCmdTable[cmdName]; ok {
			for e := server.slavelist.Front(); e != nil; e = e.Next() {
				// TODO
				// e.Value.(*SlaveSession).AsyncSendCommand(cmd)
			}
		}
	}()

	// monitor
	go server.monitorOutput(session, cmd)
}

func (server *GoRedisServer) OnUndefined(session *Session, cmd *Command) (reply *Reply) {
	return ErrorReply("Not Supported: " + cmd.String())
}

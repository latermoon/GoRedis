package goredis_server

import (
	. "../goredis"
	"../libs/levelredis"
	statlog "../libs/statlog"
	stdlog "../libs/stdlog"
	"../libs/uuid"
	"./monitor"
	"container/list"
	"errors"
	"strings"
	"sync"
)

// 版本号，每次更新都需要升级一下
const VERSION = "1.0.15"

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
	config     *Config
	// counters
	counters        *monitor.Counters
	cmdCounters     *monitor.Counters
	cmdCateCounters *monitor.Counters // 指令集统计
	// logger
	cmdMonitor    *monitor.StatusLogger
	leveldbStatus *statlog.StatLogger
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
	server.counters = monitor.NewCounters()
	server.cmdCounters = monitor.NewCounters()
	server.cmdCateCounters = monitor.NewCounters()
	return
}

func (server *GoRedisServer) Listen(host string) {
	stdlog.Printf("listen %s\n", host)
	server.RedisServer.Listen(host)
}

func (server *GoRedisServer) UID() (uid string) {
	if len(server.uid) == 0 {
		uidkey := "uid"
		server.uid = server.config.StringForKey(uidkey)
		if len(server.uid) == 0 {
			server.uid = uuid.UUID(8)
			server.config.Set(uidkey, []byte(server.uid))
		}
	}
	return server.uid
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

// for CommandHandler
func (server *GoRedisServer) On(session *Session, cmd *Command) {
	go func() {
		cmdName := strings.ToUpper(cmd.Name())
		server.cmdCounters.Get(cmdName).Incr(1)
		cate := GetCommandCategory(cmdName)
		server.cmdCateCounters.Get(string(cate)).Incr(1)
		server.cmdCateCounters.Get("total").Incr(1)

		// 同步到从库
		if _, ok := server.needSyncCmdTable[cmdName]; ok {
			for e := server.slavelist.Front(); e != nil; e = e.Next() {
				// TODO
				// e.Value.(*SlaveSession).AsyncSendCommand(cmd)
			}
		}
	}()

	// monitor
	if server.monitorlist.Len() > 0 {
		go server.monitorOutput(session, cmd)
	}
}

func (server *GoRedisServer) OnUndefined(session *Session, cmd *Command) (reply *Reply) {
	return ErrorReply("NotSupported: " + strings.ToUpper(cmd.Name()))
}

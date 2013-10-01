package goredis_server

import (
	. "../goredis"
	"./monitor"
	"./storage"
	"strings"
	"sync"
)

var (
	WrongKindReply = ErrorReply("Wrong kind opration")
)

// GoRedisServer
type GoRedisServer struct {
	CommandHandler
	RedisServer
	// 数据源
	datasource storage.DataSource
	// counters
	counters     map[string]*monitor.Counter
	counterMutex sync.Mutex
	// logger
	statusLogger *monitor.StatusLogger
	// 从库
	slaveMgr *SlaveServerManager
	// 当前实例名字
	uid string
	// 从库状态
	ReplicationInfo ReplicationInfo
}

func NewGoRedisServer() (server *GoRedisServer) {
	server = &GoRedisServer{}
	// set as itself
	server.SetHandler(server)
	// default datasource
	server.datasource = storage.NewMemoryDataSource()
	server.counters = make(map[string]*monitor.Counter)
	// slave
	server.slaveMgr = NewSlaveServerManager(server)
	server.ReplicationInfo = ReplicationInfo{}
	return
}

func (server *GoRedisServer) Listen(host string) {
	port := strings.Split(host, ":")[1]
	var e1 error
	server.datasource, e1 = storage.NewLevelDBDataSource("/tmp/goredis_" + port + ".ldb")
	if e1 != nil {
		panic(e1)
	}
	server.statusLogger = monitor.NewStatusLogger("/tmp/monitor_" + port + ".log")
	server.statusLogger.Add(monitor.NewTimeFormater("Time", 8))
	cmds := []string{"TOTAL", "GET", "SET"}
	for _, cmd := range cmds {
		server.statusLogger.Add(monitor.NewCountFormater(server.Counter(cmd), cmd, 7))
	}
	server.initUID()
	server.statusLogger.Start()
	server.RedisServer.Listen(host)
}

func (server *GoRedisServer) Counter(name string) (counter *monitor.Counter) {
	server.counterMutex.Lock()
	defer server.counterMutex.Unlock()
	var exist bool
	counter, exist = server.counters[name]
	if !exist {
		counter = monitor.NewCounter()
		server.counters[name] = counter
	}
	return
}

func (server *GoRedisServer) initUID() {
	// uuidKey := "__goredis_uuid__"
	// data, e1 := server.Storages.StringStorage.Get(uuidKey)
	// if e1 != nil {
	// 	panic(e1)
	// }
	// if data != nil {
	// 	switch data.(type) {
	// 	case string:
	// 		server.uid = data.(string)
	// 	case []byte:
	// 		server.uid = string(data.([]byte))
	// 	default:
	// 		panic("Bad UUID")
	// 	}
	// } else {
	// 	server.uid = uuid.NewV4().String()
	// 	server.Storages.StringStorage.Set(uuidKey, server.uid)
	// }
	// fmt.Println("GoRedis UUID:", server.UID())
}

func (server *GoRedisServer) UID() string {
	return server.uid
}

// for CommandHandler
func (server *GoRedisServer) On(cmd *Command, session *Session) {
	go func() {
		server.Counter(strings.ToUpper(cmd.Name())).Incr(1)
		server.Counter("TOTAL").Incr(1)
	}()
}

func (server *GoRedisServer) OnUndefined(cmd *Command, session *Session) (reply *Reply) {
	return ErrorReply("Not Supported: " + cmd.String())
}

package goredis_server

// GoRedis核心类
import (
	. "GoRedis/goredis"
	"GoRedis/libs/counter"
	"GoRedis/libs/levelredis"
	"GoRedis/libs/statlog"
	"GoRedis/libs/stdlog"
	"GoRedis/libs/uuid"
	"errors"
	"os"
	"reflect"
	"strings"
	"time"
)

// 版本号，每次更新都需要升级一下
const VERSION = "1.0.45"

var (
	WrongKindError = errors.New("Wrong kind opration")
	WrongKindReply = ErrorReply(WrongKindError)
)

var goredisPrefix string = "__goredis:"

var (
	slowexec = 30 // ms
	slowlog  = stdlog.Log("slow")
	errlog   = stdlog.Log("err")
)

// GoRedisServer
type GoRedisServer struct {
	ServerHandler
	RedisServer
	// 数据源
	directory  string
	levelRedis *levelredis.LevelRedis
	config     *Config
	// counters
	counters        *counter.Counters
	cmdCounters     *counter.Counters
	cmdCateCounters *counter.Counters // 指令集统计
	// logger
	cmdMonitor    *statlog.StatLogger
	leveldbStatus *statlog.StatLogger
	// 从库
	uid      string        // 实例id
	syncmgr  *SyncManager  // as master
	slavemgr *SlaveManager // as slave
	// monitor
	monmgr *MonManager
	// 缓存处理函数，减少relect次数
	methodCache map[string]reflect.Value
	// 指令队列，异步处理统计、从库、monitor输出
	cmdChan chan *Command
	// exit
	sigs     chan os.Signal
	quitdone chan bool // 准备好退出
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
	server.methodCache = make(map[string]reflect.Value)
	server.cmdChan = make(chan *Command, 1000)
	go server.processCommandChan()
	server.syncmgr = NewSyncManager()
	server.slavemgr = NewSlaveManager()
	server.monmgr = NewMonManager()
	// default datasource
	server.directory = directory
	// counter
	server.counters = counter.NewCounters()
	server.cmdCounters = counter.NewCounters()
	server.cmdCateCounters = counter.NewCounters()
	return
}

func (server *GoRedisServer) Listen(host string) error {
	stdlog.Printf("listen %s\n", host)
	return server.RedisServer.Listen(host)
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

// ServerHandler.SessionOpened()
func (server *GoRedisServer) SessionOpened(session *Session) {
	server.counters.Get("connection").Incr(1)
	stdlog.Println("connection accepted from", session.RemoteAddr())
}

// ServerHandler.SessionClosed()
func (server *GoRedisServer) SessionClosed(session *Session, err error) {
	server.counters.Get("connection").Incr(-1)
	stdlog.Println("end connection", session.RemoteAddr(), err)
}

// ServerHandler.On()
// 由GoRedis协议层触发，通过反射调用OnGET/OnSET等方法
func (server *GoRedisServer) On(session *Session, cmd *Command) (reply *Reply) {
	if err := verifyCommand(cmd); err != nil {
		errlog.Printf("[%s] bad command %s\n", session.RemoteAddr(), cmd)
		return ErrorReply(err)
	}
	// invoke & time
	begin := time.Now()
	reply = server.invokeCommandHandler(session, cmd)
	elapsed := time.Now().Sub(begin)

	if elapsed.Nanoseconds() > int64(time.Millisecond*time.Duration(slowexec)) {
		slowlog.Printf("[%s] exec %0.2f ms [%s]\n", session.RemoteAddr(), elapsed.Seconds()*1000, cmd)
	}
	// see processCommandChan()
	server.cmdChan <- cmd
	return
}

// 异步串行处理
func (server *GoRedisServer) processCommandChan() {
	for {
		cmd := <-server.cmdChan
		cmdName := strings.ToUpper(cmd.Name())

		server.incrCommandCounter(cmdName)

		// 从库
		if server.syncmgr.Count() > 0 && needSync(cmdName) {
			server.syncmgr.BroadcastCommand(cmd)
		}

		// monitor
		if server.monmgr.Count() > 0 {
			server.monmgr.BroadcastCommand(cmd)
		}
	}
}

// 指令计数器
func (server *GoRedisServer) incrCommandCounter(cmdName string) {
	server.cmdCounters.Get(cmdName).Incr(1)
	cate := commandCategory(cmdName)
	server.cmdCateCounters.Get(string(cate)).Incr(1)
	server.cmdCateCounters.Get("total").Incr(1)
}

// 首先搜索"On+大写NAME"格式的函数，存在则调用，不存在则调用OnUndefined
// OnGET(cmd *Command) (reply *Reply)
// OnGET(session *Session, cmd *Command) (reply *Reply)
func (server *GoRedisServer) invokeCommandHandler(session *Session, cmd *Command) (reply *Reply) {
	cmdName := strings.ToUpper(cmd.Name())
	// 从Cache取出处理函数
	method, exists := server.methodCache[cmdName]
	if !exists {
		method = reflect.ValueOf(server).MethodByName("On" + cmdName)
		server.methodCache[cmdName] = method
	}

	if method.IsValid() {
		// 可以调用两种接口
		// method = OnXXX(cmd *Command) (reply *Reply)
		// method = OnXXX(session *Session, cmd *Command) (reply *Reply)
		var in []reflect.Value
		if method.Type().NumIn() == 1 {
			in = []reflect.Value{reflect.ValueOf(cmd)}
		} else {
			in = []reflect.Value{reflect.ValueOf(session), reflect.ValueOf(cmd)}
		}
		callResult := method.Call(in)
		reply = callResult[0].Interface().(*Reply)
	} else {
		reply = server.OnUndefined(session, cmd)
	}

	return
}

func (server *GoRedisServer) OnUndefined(session *Session, cmd *Command) (reply *Reply) {
	return ErrorReply("NotSupported: " + cmd.String())
}

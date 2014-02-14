package goredis_server

import (
	"./monitor"
	. "GoRedis/goredis"
	"GoRedis/libs/levelredis"
	"GoRedis/libs/statlog"
	"GoRedis/libs/stdlog"
	"GoRedis/libs/uuid"
	"container/list"
	"errors"
	"reflect"
	"strings"
	"sync"
	"time"
)

// 版本号，每次更新都需要升级一下
const VERSION = "1.0.21"

var (
	WrongKindError = errors.New("Wrong kind opration")
	WrongKindReply = ErrorReply(WrongKindError)
)

var goredisPrefix string = "__goredis:"

var (
	slowexec = 30 // ms
	slowlog  = stdlog.Log("slow")
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
	counters        *monitor.Counters
	cmdCounters     *monitor.Counters
	cmdCateCounters *monitor.Counters // 指令集统计
	// logger
	cmdMonitor    *monitor.StatusLogger
	leveldbStatus *statlog.StatLogger
	// 从库
	uid       string // 实例id
	slavelist *list.List
	// monitor
	monitorlist  *list.List
	monitorMutex sync.Mutex
	// 缓存处理函数，减少relect次数
	methodCache map[string]reflect.Value
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
	// default datasource
	server.directory = directory
	server.slavelist = list.New()
	server.monitorlist = list.New()
	// counter
	server.counters = monitor.NewCounters()
	server.cmdCounters = monitor.NewCounters()
	server.cmdCateCounters = monitor.NewCounters()
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
		return ErrorReply(err)
	}
	// slowlog
	begin := time.Now()
	// invoke
	reply = server.invokeCommandHandler(session, cmd)
	elapsed := time.Now().Sub(begin)
	if elapsed.Nanoseconds() > int64(time.Millisecond*time.Duration(slowexec)) {
		slowlog.Printf("[%s] exec %0.2f ms [%s]\n", session.RemoteAddr(), elapsed.Seconds()*1000, cmd)
	}

	// monitor
	if server.monitorlist.Len() > 0 {
		go server.monitorOutput(session, cmd)
	}

	// 这里要注意并发
	go func() {
		cmdName := strings.ToUpper(cmd.Name())
		server.cmdCounters.Get(cmdName).Incr(1)
		cate := commandCategory(cmdName)
		server.cmdCateCounters.Get(string(cate)).Incr(1)
		server.cmdCateCounters.Get("total").Incr(1)

		// 同步到从库
		if needSync(cmdName) {
			for e := server.slavelist.Front(); e != nil; e = e.Next() {
				sc := e.Value.(*SyncClient)
				sc.SendCommand(cmd)
			}
		}
	}()

	return
}

// 首先搜索"On+大写NAME"格式的函数，存在则调用，不存在则调用On
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

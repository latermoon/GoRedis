package goredis_server

// GoRedis核心类
import (
	. "GoRedis/goredis"
	"GoRedis/libs/counter"
	"GoRedis/libs/levelredis"
	"GoRedis/libs/stdlog"
	"GoRedis/libs/uuid"
	"container/list"
	"errors"
	"fmt"
	"os"
	"reflect"
	"runtime/debug"
	"sync"
	"time"
)

// 版本号，每次更新都需要增加
const VERSION = "1.0.72"
const PREFIX = "__goredis:"

var (
	WrongKindError = errors.New("Wrong kind opration")
	WrongKindReply = ErrorReply(WrongKindError)
)

var (
	slowexec = float64(30) // ms
	slowlog  = stdlog.Log("slow")
)

// GoRedisServer
type GoRedisServer struct {
	ServerHandler
	RedisServer
	// 数据源
	opt        *Options // 选项
	levelRedis *levelredis.LevelRedis
	config     *Config
	// counters
	counters        *counter.Counters
	cmdCounters     *counter.Counters
	cmdCateCounters *counter.Counters // 指令集统计
	execCounters    *counter.Counters //指令执行时间计数器
	// info
	info *Info
	// 从库
	uid       string          // 实例id
	syncmgr   *SessionManager // as master
	slavemgr  *SessionManager // as slave
	synclog   *SyncLog
	aofwriter *AOFWriter
	// monitor
	sessmgr     *SessionManager // all sessions
	monmgr      *SessionManager
	methodCache map[string]reflect.Value // 缓存处理函数，减少relect次数
	cmdChan     chan *Command            // 指令队列，异步处理统计、从库、monitor输出
	rwlock      sync.RWMutex
	rwwait      sync.WaitGroup
	// exit
	sigs        chan os.Signal
	closing     bool       // 准备退出
	closingFunc *list.List // 退出执行函数FIFO
}

// server := NewGoRedisServer(opt)
// server.Init()
// server.Listen(host)
func NewGoRedisServer(opt *Options) (server *GoRedisServer) {
	server = &GoRedisServer{}
	server.opt = opt
	// set as itself
	server.SetHandler(server)
	server.methodCache = make(map[string]reflect.Value)
	server.cmdChan = make(chan *Command, 1000)
	server.closingFunc = list.New()
	go server.processCommandChan()
	server.monmgr = NewSessionManager()
	server.syncmgr = NewSessionManager()
	server.slavemgr = NewSessionManager()
	server.sessmgr = NewSessionManager()
	server.info = NewInfo(server)
	// counter
	server.counters = counter.NewCounters()
	server.cmdCounters = counter.NewCounters()
	server.cmdCateCounters = counter.NewCounters()
	server.execCounters = counter.NewCounters()
	return
}

func (server *GoRedisServer) Listen() error {
	addr := fmt.Sprintf("%s:%d", server.opt.Host(), server.opt.Port())
	stdlog.Printf("listen %s\n", addr)
	return server.RedisServer.Listen(addr)
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
	server.sessmgr.Put(session.RemoteAddr().String(), session)
	stdlog.Println("connection accepted from", session.RemoteAddr())
}

// ServerHandler.SessionClosed()
func (server *GoRedisServer) SessionClosed(session *Session, err error) {
	server.counters.Get("connection").Incr(-1)
	server.sessmgr.Remove(session.RemoteAddr().String())
	stdlog.Println("end connection", session.RemoteAddr(), err)
}

func (server *GoRedisServer) ExceptionCaught(err error) {
	stdlog.Printf("exception %s\n", err)
	stdlog.Println(string(debug.Stack()))
}

// ServerHandler.On()
// 由GoRedis协议层触发，通过反射调用OnGET/OnSET等方法
func (server *GoRedisServer) On(session *Session, cmd *Command) (reply *Reply) {
	// invoke & time
	begin := time.Now()

	// suspend & resume
	server.rwlock.Lock()
	server.rwlock.Unlock()

	cmd.SetAttribute(C_SESSION, session)

	// varify command
	if err := verifyCommand(cmd); err != nil {
		stdlog.Printf("[%s] bad command %s\n", session.RemoteAddr(), cmd)
		return ErrorReply(err)
	}

	// invoke
	reply = server.invokeCommandHandler(session, cmd)

	elapsed := time.Now().Sub(begin)
	cmd.SetAttribute(C_ELAPSED, elapsed)

	// async: counter/sync/monitor
	server.rwwait.Add(1)
	server.cmdChan <- cmd

	return
}

// 挂起指令处理
func (server *GoRedisServer) Suspend() {
	server.rwlock.Lock() // 锁定On(...)入口
	server.rwwait.Wait() // 等待队列清空
}

// 唤醒指令处理
func (server *GoRedisServer) Resume() {
	server.rwlock.Unlock() // 解锁
}

// 异步串行处理
func (server *GoRedisServer) processCommandChan() {
	for {
		cmd := <-server.cmdChan
		cmdName := cmd.Name()

		// last cmd
		session := cmd.GetAttribute(C_SESSION).(*Session)
		session.SetAttribute(S_LAST_COMMAND, cmdName)

		server.incrCommandCounter(cmdName)

		// 从库
		if server.synclog.IsEnabled() && needSync(cmdName) {
			server.synclog.Write(cmd.Bytes())
		}

		// monitor
		if server.monmgr.Len() > 0 {
			server.broadcastMonitor(cmd)
		}

		// slowlog
		server.calcExecTime(cmd)

		server.rwwait.Done()
	}
}

// 指令计数器
func (server *GoRedisServer) incrCommandCounter(cmdName string) {
	server.cmdCounters.Get(cmdName).Incr(1)
	cate := commandCategory(cmdName)
	server.cmdCateCounters.Get(string(cate)).Incr(1)
	server.cmdCateCounters.Get("total").Incr(1)
}

// 向monitor clients广播
func (server *GoRedisServer) broadcastMonitor(cmd *Command) {
	server.monmgr.Enumerate(func(i int, key string, val interface{}) {
		c := val.(*MonClient)
		c.Send(cmd)
	})
}

// 统计指令耗时
func (server *GoRedisServer) calcExecTime(cmd *Command) {
	elapsed := cmd.GetAttribute(C_ELAPSED).(time.Duration)
	msec := elapsed.Seconds() * 1000
	switch {
	case msec < 1:
		server.execCounters.Get("<1ms").Incr(1)
	case msec >= 1 && msec <= 5:
		server.execCounters.Get("1-5ms").Incr(1)
	case msec > 5 && msec <= 10:
		server.execCounters.Get("6-10ms").Incr(1)
	case msec > 10 && msec <= 30:
		server.execCounters.Get("11-30ms").Incr(1)
	case msec > 30:
		server.execCounters.Get(">30ms").Incr(1)
	}
	if msec > slowexec {
		session := cmd.GetAttribute(C_SESSION).(*Session)
		slowlog.Printf("[%s] exec %0.2f ms [%s]\n", session.RemoteAddr(), msec, cmd)
	}
}

// 首先搜索"On+大写NAME"格式的函数，存在则调用，不存在则调用OnUndefined
// OnGET(cmd *Command) (reply *Reply)
// OnGET(session *Session, cmd *Command) (reply *Reply)
func (server *GoRedisServer) invokeCommandHandler(session *Session, cmd *Command) (reply *Reply) {
	cmdName := cmd.Name()
	method, exists := server.methodCache[cmdName]
	if !exists {
		method = reflect.ValueOf(server).MethodByName("On" + cmdName)
		server.methodCache[cmdName] = method
	}

	if method.IsValid() {
		var in []reflect.Value
		if method.Type().NumIn() == 1 {
			in = []reflect.Value{reflect.ValueOf(cmd)}
		} else {
			in = []reflect.Value{reflect.ValueOf(session), reflect.ValueOf(cmd)}
		}
		callResult := method.Call(in)
		if callResult[0].Interface() != nil {
			reply = callResult[0].Interface().(*Reply)
		}
	} else {
		reply = server.OnUndefined(session, cmd)
	}

	return
}

func (server *GoRedisServer) OnUndefined(session *Session, cmd *Command) (reply *Reply) {
	return ErrorReply("NotSupported: " + cmd.String())
}

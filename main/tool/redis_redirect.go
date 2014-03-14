package main

// 通过MONITOR将Redis指令转发到另外的Redis或GoRedis
// TODO 过程式代码优化
import (
	. "GoRedis/goredis"
	"GoRedis/libs/redis_tool"
	"GoRedis/libs/stdlog"
	"flag"
	"fmt"
	"github.com/latermoon/redigo/redis"
	"runtime"
	"strings"
	"sync"
	"time"
)

// 写操作指令
var syncCmdlist = "DEL,EXPIRE,PERSIST,PEXPIRE,PEXPIREAT,RENAME,RENAMENX,SORT,APPEND,DECR,DECRBY,INCR,INCRBY,INCRBYFLOAT,MSET,MSETNX,PSETEX,SET,SETBIT,SETEX,SETNX,SETRANGE,HDEL,HINCRBY,HINCRBYFLOAT,HMSET,HSET,HSETNX,BLPOP,BRPOP,BRPOPLPUSH,LINSERT,LPOP,LPUSH,LPUSHX,LREM,LSET,LTRIM,RPOP,RPOPLRUSH,RPUSH,RPUSHX,SADD,SDIFFSTORE,SINTERSTORE,SMOVE,SPOP,SREM,SUNIONSTORE,ZADD,ZINCRBY,ZINTERSTORE,ZREM,ZREMRANGEBYRANK,ZREMRANGEBYSCORE,ZREVRANK,ZUNIONSTORE"

// 跳过指令
var ignoreCmdList = "DUMP,KEYS,MIGRATE,MOVE,OBJECT,RESTORE,SCAN,EVAL,EVALSHA,SCRIPT,DISCARD,EXEC,MULTI,UNWATCH,WATCH,PSUBSCRIBE,PUBSUB,PUBLISH,PUNSUBSCRIBE,SUBSCRIBE,UNSUBSCRIBE,BGREWRITEAOF,BGSAVE,CLIENT,CONFIG,DBSIZE,DEBUG,FLUSHALL,FLUSHDB,INFO,LASTSAVE,MONITOR,SAVE,SHUTDOWM,SLAVEOF,SLOWLOG,SYNC,TIME"

var needSync = map[string]bool{}
var ignoreSync = map[string]bool{}

var desthost string
var pools map[string]*redis.Pool
var mu sync.Mutex
var mode string
var buffer chan *Command

func init() {
	runtime.GOMAXPROCS(4)
	pools = make(map[string]*redis.Pool)
	buffer = make(chan *Command, 100*10000) // 最大缓存100w

	for _, cmd := range strings.Split(syncCmdlist, ",") {
		needSync[cmd] = true
	}
	for _, cmd := range strings.Split(ignoreCmdList, ",") {
		ignoreSync[cmd] = true
	}

	stdlog.SetPrefix(func() string {
		t := time.Now()
		return fmt.Sprintf("[%d-%02d-%02d %02d:%02d:%02d] ", t.Year(), t.Month(), t.Day(), t.Hour(), t.Minute(), t.Second())
	})
}

func main() {
	srcptr := flag.String("src", "", "source host")
	destptr := flag.String("dest", "", "dest host")
	modePtr := flag.String("mode", "", "r/w/rw")
	flag.Parse()

	if len(*srcptr) == 0 || len(*destptr) == 0 {
		stdlog.Println("must set -src or -dest")
		return
	}
	desthost = *destptr
	mode = *modePtr
	if len(mode) == 0 {
		stdlog.Println("must set -mode [r|w|rw]")
		return
	}

	go runloop()

	r := redis_tool.NewMonitorReader(*srcptr)
	r.DidRecvCommand = recvCommand // bind
	err := r.Connect()
	if err != nil {
		panic(err)
	}
}

func recvCommand(cmd *Command, prefix string) {
	cmdName := strings.ToUpper(cmd.Name())
	if ignoreSync[cmdName] {
		return
	}
	if strings.Contains(mode, "r") {
		if !needSync[cmdName] {
			buffer <- cmd
		}
	}
	if strings.Contains(mode, "w") {
		if needSync[cmdName] {
			buffer <- cmd
		}
	}
}

func runloop() {
	for {
		cmd := <-buffer
		ok := false
		for !ok {
			ok = redirectToGoRedis(cmd)
			if !ok {
				time.Sleep(time.Millisecond * 1000)
			}
		}
	}
}

var total int64 = 0

func redirectToGoRedis(cmd *Command) (ok bool) {
	pool := GetRedisPool(desthost)
	conn := pool.Get()
	defer conn.Close()
	args := make([]interface{}, 0, len(cmd.Args)-1)
	for i := 1; i < len(cmd.Args); i++ {
		args = append(args, cmd.Args[i])
	}
	_, err := conn.Do(cmd.StringAtIndex(0), args...)
	if err != nil {
		// io.EOF or "connection refused"
		stdlog.Println("ERR", len(buffer), total, cmd, err)
		return false
	}
	total++
	stdlog.Println(len(buffer), total, cmd)
	return true
}

func GetRedisPool(host string) (pool *redis.Pool) {
	mu.Lock()
	defer mu.Unlock()
	var exist bool
	pool, exist = pools[host]
	if !exist {
		pool = &redis.Pool{
			MaxIdle:     100,
			IdleTimeout: 240 * time.Second,
			Dial: func() (redis.Conn, error) {
				c, err := redis.Dial("tcp", host)
				return c, err
			},
		}
		pools[host] = pool
	}
	return
}

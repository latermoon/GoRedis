package shardredis

import (
	. "GoRedis/goredis"
	"GoRedis/libs/redigo/redis"
	"fmt"
	"strings"
	"time"
)

// 写操作指令
var syncCmdlist = "DEL,EXPIRE,PERSIST,PEXPIRE,PEXPIREAT,RENAME,RENAMENX,SORT,APPEND,DECR,DECRBY,INCR,INCRBY,INCRBYFLOAT,MSET,MSETNX,PSETEX,SET,SETBIT,SETEX,SETNX,SETRANGE,HDEL,HINCRBY,HINCRBYFLOAT,HMSET,HSET,HSETNX,BLPOP,BRPOP,BRPOPLPUSH,LINSERT,LPOP,LPUSH,LPUSHX,LREM,LSET,LTRIM,RPOP,RPOPLRUSH,RPUSH,RPUSHX,SADD,SDIFFSTORE,SINTERSTORE,SMOVE,SPOP,SREM,SUNIONSTORE,ZADD,ZINCRBY,ZINTERSTORE,ZREM,ZREMRANGEBYRANK,ZREMRANGEBYSCORE,ZREVRANK,ZUNIONSTORE"

// 跳过指令
var ignoreCmdList = "DUMP,KEYS,MIGRATE,MOVE,OBJECT,RESTORE,SCAN,EVAL,EVALSHA,SCRIPT,DISCARD,EXEC,MULTI,UNWATCH,WATCH,PSUBSCRIBE,PUBSUB,PUBLISH,PUNSUBSCRIBE,SUBSCRIBE,UNSUBSCRIBE,BGREWRITEAOF,BGSAVE,CLIENT,CONFIG,DBSIZE,DEBUG,FLUSHALL,FLUSHDB,INFO,LASTSAVE,MONITOR,SAVE,SHUTDOWM,SLAVEOF,SLOWLOG,SYNC,TIME"

var needSync = map[string]bool{}
var ignoreSync = map[string]bool{}

func init() {
	for _, cmd := range strings.Split(syncCmdlist, ",") {
		needSync[cmd] = true
	}
	for _, cmd := range strings.Split(ignoreCmdList, ",") {
		ignoreSync[cmd] = true
	}
}

// 使用monitor获取数据，将读写转发到目的地
type MonitorRedirect struct {
	src    string
	dest   string
	mode   string
	buffer chan *Command
	pool   *redis.Pool
}

func NewMonitorRedirect(src, dest, mode string) (m *MonitorRedirect) {
	m = &MonitorRedirect{
		src:    src,
		dest:   dest,
		mode:   mode,
		buffer: make(chan *Command, 100*10000),
	}
	m.pool = RedisPool(m.dest)
	return
}

func (m *MonitorRedirect) Start() (err error) {
	reader := NewMonitorReader(m.src)
	reader.DidRecvCommand = m.didRecvCommand
	go m.runloop()
	err = reader.Connect()
	return
}

func (m *MonitorRedirect) runloop() {
	for {
		cmd := <-m.buffer
		// fmt.Println(cmd)
		var err error
		for err != nil {
			err = m.redirect(cmd)
			if err != nil {
				fmt.Println(err, cmd)
				time.Sleep(time.Millisecond * 1000)
			}
		}
	}
}

// 转发
func (m *MonitorRedirect) redirect(cmd *Command) (err error) {
	conn := m.pool.Get()
	defer conn.Close()

	args := make([]interface{}, cmd.Len()-1)
	for i := 1; i < cmd.Len(); i++ {
		args[i-1] = cmd.Args()[i]
	}
	_, err = conn.Do(cmd.Name(), args...)
	return
}

// 处理收到的信息
func (m *MonitorRedirect) didRecvCommand(cmd *Command, prefix string) {
	cmdName := strings.ToUpper(cmd.Name())
	if ignoreSync[cmdName] {
		return
	}

	if m.allowRead() {
		if !needSync[cmdName] {
			m.buffer <- cmd
		}
	}

	if m.allowWrite() {
		if needSync[cmdName] {
			m.buffer <- cmd
		}
	}
}

func (m *MonitorRedirect) allowRead() bool {
	return strings.Contains(m.mode, "r")
}

func (m *MonitorRedirect) allowWrite() bool {
	return strings.Contains(m.mode, "w")
}

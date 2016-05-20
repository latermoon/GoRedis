package main

import (
	. "GoRedis/goredis"
	"GoRedis/libs/redis_tool"
	"flag"
	"fmt"
	"github.com/latermoon/GoRedis/libs/stdlog"
	"github.com/latermoon/redigo/redis"
	"runtime"
	"strings"
	"sync"
	"time"
)

var desthost string
var pools map[string]*redis.Pool
var poolmu sync.Mutex

func init() {
	runtime.GOMAXPROCS(8)
	pools = make(map[string]*redis.Pool)
}

func main() {
	srcptr := flag.String("host", "", "host")
	flag.Parse()

	if len(*srcptr) == 0 || len(*destptr) == 0 {
		stdlog.Println("bad src or dest")
		return
	}
	desthost = *destptr

	r := redis_tool.NewMonitorReader(*srcptr)
	r.DidRecvCommand = recvCommand // bind
	err := r.Connect()
	if err != nil {
		panic(err)
	}
}

func recvCommand(cmd *Command, prefix string) {
	// only
	// Command <SADD user:[momoid]:replyed [remoteid]>
	if len(cmd.Args) != 3 {
		return
	}
	if cmd.StringAtIndex(0) != "SADD" {
		return
	}
	if !strings.Contains(cmd.StringAtIndex(1), "replyed") {
		return
	}

	// 将SADD转换为ZADD存放到GoRedis
	key, _ := cmd.ArgAtIndex(1)
	score := []byte(fmt.Sprint(time.Now().UnixNano() / 1000000))
	member, _ := cmd.ArgAtIndex(2)

	zcmd := NewCommand([]byte("ZADD"), key, score, member)
	redirectToGoRedis(zcmd)
}

func redirectToGoRedis(cmd *Command) {
	pool := GetRedisPool(desthost)
	conn := pool.Get()
	defer conn.Close()
	args := make([]interface{}, 0, len(cmd.Args)-1)
	for i := 1; i < len(cmd.Args); i++ {
		args = append(args, cmd.Args[i])
	}
	reply, err := conn.Do(cmd.StringAtIndex(0), args...)
	if err != nil {
		panic(err)
	}
	stdlog.Println(reply, cmd)
}

func GetRedisPool(host string) (pool *redis.Pool) {
	poolmu.Lock()
	defer poolmu.Unlock()
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

git clone -b GoRedisDev https://github.com/latermoon/GoRedis.git GoRedisDev
git fetch


### goredis-server config

GoRedis开发进展

和官方版一致的地方：
1、可以和官方redis互为主从，支持大部分常用指令集：string、hash、list、set、sortedset
2、同步持久化的SET平均6w+/s、GET平均5w/s，增加LRUCache后，可以达到10w/s左右（大致数据，看开的CPU情况）
3、因为leveldb的特点，最适合用SSD服务器

扩展特性：
1、快速启动，持久层使用LevelDB，重启不会丢数据，即时启动，不需要reload rdb
2、增量同步，GoRedis主从情况下，从库断开连接后，再次连上可以增量（参考mongodb的做法）
3、增加aof_push指令，也是基于leveldb，可以实现海量日志存储，可以用来做消息备份，用户资料历史

```
__goredis:uid = 4faedbb8
__goredis:version = 2
__goredis:slaves = [d03b4f61, c8fb5668]
__goredis:slave:d03b4f61:info = {
	uid: "d03b4f61",
	aof_start: 1004
	aof_end: 1008
	creation_time: "2013-10-05 12:13:07"
	last_connect_time: "2013-10-05 12:13:07"
}

```

#### aof list

\[prefix]:_start = 1004 (int64)
\[prefix]:_end = 1008 (int64)
\[prefix]:idx:1004 = hello ([]byte)
\[prefix]:idx:1005 = hello
\[prefix]:idx:1006 = hello
\[prefix]:idx:1007 = hello
\[prefix]:idx:1008 = hello

ds

#### log
cmd.log 
sync.log
stdout.log
[2013-10-06 01:20:08] 
stderr.log

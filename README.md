GoRedis
=======

### RedisServer Implemented by Go
#### 说明
	1、围绕Redis协议衍生出的数据处理框架
	2、包括goredis和goredis-server，前者是一个业务无关的redisserver协议层，后者是全功能的定制redisserver

#### goredis-server
	和官方Redis一致的地方：
	1、可以和官方redis互为主从，支持大部分常用指令集：string、hash、list、set、sortedset
	2、同步持久化的SET/GET平均6w+/s （SSD服务器）
	3、因为leveldb的特点，最适合用SSD服务器

	扩展特性：
	1、快速启动，持久层使用LevelDB，重启不会丢数据，即时启动，不需要reload rdb
	2、增量同步，GoRedis主从情况下，从库断开连接后，再次连上可以增量（参考mongodb的做法）
	3、Hash/Set/List/SortedSet也是基于leveldb的特点设计，可以实现海量日志存储而不消耗内存
	4、MultiSlave，一个GoRedis可以同时作为多个Redis、GoRedis的从库，适用于不同key的海量备份

#### TODO
	1、精细修正命名和注释、日志等维护性代码
	2、修正全部需要分配buf的地方，监控过大内存的分配
	3、校对全部有error输出的地方是否处理正确
	4、校对全部由用户传入的数据格式是否合法

#### Install:

	git clone github.com/latermoon/GoRedis
	cd main/
	go build goredis-server.go

#### Run:

	./goredis-server -procs 8 -p 1602




GoRedis
=======

*请注意：GoRedis处于beta阶段，为陌陌公司内部测试，请保持关注，等待1.0发布*

### RedisServer Implemented by Go
#### 说明
	1、使用rocksdb作为存储层的RedisServer，数据不消耗内存，保持较高性能的情况下，同时获得海量存储特性
	2、可以和官方redis互为主从，支持常用指令集，全部redis指令会转换为rocksdb操作
	3、因为rocksdb的特点，最适合用SSD服务器，SET/GET平均7w+/s

	扩展特性：
	1、快速启动，持久层使用rocksdb，重启不会丢数据，即时启动，不需要reload rdb
	2、增量同步，GoRedis主从情况下，从库断开连接后，再次连上可以增量同步，适合海量存储和跨机房同步
	3、Hash/Set/List/SortedSet也是基于rocksdb的特点设计，可以实现海量日志存储而不消耗内存
	4、MultiSlave，GoRedis之间可以一主多从和一从多主

#### Install:
	先安装rocksdb，复制代码，编译GoRedis
	git clone github.com/latermoon/GoRedis
	sh install.sh
	sh build.sh

#### Run:
	cd /home/server/goredis/bin/
	./goredis-server -procs 8 -p 1602



### GoRedis日志查看

GoRedis带有丰富的日志输出，详细描述每种指令的执行次数、慢查询、rocksdb操作等。

日志均在/data/goredis_[port]目录下。

#### 标准输出
1、stdout.log

	[2014-03-04 19:12:52] server init ...
	[2014-03-04 19:12:52] init uid 1a819a66
	[2014-03-04 19:12:52] listen :1604
	[2014-03-04 19:12:53] connection accepted from 127.0.0.1:50822
	[2014-03-05 00:14:58] end connection 127.0.0.1:50822 EOF

#### 指令操作数

1、cmd.log:

	    time  total    key string   hash   list    set   zset   conn
	11:03:11   1017      0      0      0    176      0    841      0
	11:03:12   1128      0      0      0    191      0    937      0
	11:03:13   1188      0      0      0    194      0    994      0

2、cmd.string.log

    	time     GET     SET    MGET    MSET    INCR    DECR  INCRBY  DECRBY
	11:06:20       0       0       0       0       0       0       0       0
	11:06:21       0       0       0       0       0       0       0       0
	11:06:22       0       0       0       0       0       0       0       0

3、cmd.hash.log/cmd.set.log/cmd.list.log/cmd.zset.log

#### rocksdb指令数
	
1、leveldb.io.log

	    time       get       set     batch      enum       del   lru_hit  lru_miss
	11:08:40      1331         0      1912       833         0      1514       499
	11:08:41      1496         0      2180       954         0      1734       544
	11:08:42      1233         0      1776       777         0      1410       462
	11:08:43      1439         0      2064       901         0      1630       545
	11:08:44      1259         0      1818       796         0      1439       467

#### 其它日志

1、慢查询 slow.log

	[2014-03-05 11:08:50] [10.80.100.216:64194] exec 25.59 ms [ZADD user:74402733:stamp 139398 wtw31f]
	[2014-03-05 11:08:50] [10.80.101.166:20000] exec 25.67 ms [ZINCRBY user:38275013:code 1.0 wssp90]
	[2014-03-05 11:09:39] [10.80.101.166:55159] exec 20.62 ms [ZADD user:39565803:stamp 188979 wqrcxr]

2、从库同步 sync_[host:port].log

	    time     raw     cmd             seq
	11:13:45       0    1082         4788124
	11:13:46       0     970         4789094
	11:13:47       0    1095         4790189
	11:13:48       0    1014         4791203

GoRedis release notes
=====================

**GoRedis 1.0.72** @2014.5.28

* [Fix] 重要，修正HSET不更新问题

**GoRedis 1.0.70** @2014.4.16

* [Feature] 优化LRANGE性能，从多次RAW_GET改为单次Enum
* [Coding] 简化启动函数代码
* [Fix] 恢复logpath默认路径为/data

**SlaveOf 1.0.4** @2014.4.11

* [Feature] 支持配置dbpath

**Proxy 1.0.4** @2014.3.30

* [Feature] 实现基本功能


**GoRedis 1.0.69** @2014.4.10

* [Fix] 修正PrefixEnumerate时全部遍历没有quit
* [ADD]	启动参数新增logpath、datapath用于自定义数据和Log路径


**GoRedis 1.0.68** @2014.3.30

* [Fix] 修正SADD返回值
* [Fix] 修正HMGET返回值
* [Fix] 修正ZADD返回值
* [Fix] 修正Session.ReadReply兼容性

**GoRedis 1.0.67** @2014.3.29

* [Fix] 修正ZRANGEBYSCORE/ZREVRANGEBYSCORE/OnZREMRANGEBYSCORE
* [Feature] 大量代码风格修正

**GoRedis 1.0.66** @2014.3.28

* [Fix] make release rocksdb, defined NDEBUG

**GoRedis 1.0.64** @2014.3.28

* [Fix] 增加DeferClosing
* [Fix] NoCache, Smaller BlockSize

**GoRedis 1.0.63** @2014.3.27

* [Fix] 修正keynext指令

**GoRedis 1.0.62** @2014.3.19

* [Feature] 调整主从同步指令，与旧版不兼容
* [Fix] 补全AOF指令

**GoRedis 1.0.61** @2014.3.14

* [Feature] 实现AOF

**GoRedis 1.0.60** @2014.3.14

* [FIX] db.Close时延时不破坏synclog操作
* [FIX] 修正Delete重构引入的bug
* [FIX] 对zset、list、hash等使用RWMutex读写锁

**GoRedis 1.0.58** @2014.3.13

* [Feature] 修正代码细节，正确显示主从端口
* [FIX] MSET的bug，用到该指令的实例需要升级

**GoRedis 1.0.56** @2014.3.12

* [Feature] 增加参数启动作为从库

**GoRedis 1.0.55** @2014.3.11

* [Feature] 增加exec.time.log记录指令性能
* [FIX] 主从连接管理
* [FIX] 大量代码简化

**GoRedis 1.0.49** @2014.3.4

* [Feature] 实现基于SEQ的增量主从同步

**GoRedis 1.0.38** @2014.2.25

* [FIX] 实现GoRedis之间的主从同步

**GoRedis 1.0.37** @2014.2.24 

* [FIX] 修复hdel死锁bug，线上版本必须全部升级


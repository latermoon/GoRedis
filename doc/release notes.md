GoRedis release notes
=====================

**GoRedis 1.0.57** @2014.3.13

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


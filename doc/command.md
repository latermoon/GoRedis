### GoRedis指令大全

#### 指令的IO消耗

和官方Redis基于内存的操作不一样，每个GoRedis指令会被转换为一个或一组rocksdb操作，理解每个指令的IO消耗对评估具体GoRedis的性能非常重要。

rocksdb操作包含Get/Set/Del/Enum，我们使用G(1)表示一次rocksdb的Get操作，同理S(2)表示两次Set操作，D、E代表Del和Enum。

下面的性能指标在2CPUx6 HT以及SLC SSD下测试，并且rocksdb的LRU对性能也有明显影响，仅供对比参考。

**当前GoRedis仅支持下列指令**

### Strings

指令 | IO | 性能 | 说明
---- | ---- | ---- | ----
GET | G(1) | 7w/s |
SET | S(1) | 6w/s | 
MGET | G(n) | | 
MSET | S(n) | | 
INCR | G(1) S(1) | | 
INCRBY | G(1) S(1) | | 
DECR | G(1) S(1) | | 
DECRBY | G(1) S(1) | | 

### Hash
指令 | IO | 性能 | 说明
---- | ---- | ---- | ----
HGET | G(1) |  |
HSET | S(2) |  | 
HGETALL | E(1) | | 
MSET | S(n) | | 
HMGET | G(n) | | 
HMSET | S(n) S(1) | | 
HLEN | E(1) | | 
HDEL | G(n) D(n) E(1) D(1) | | 删除hash的元素成本很高，需要Get判断是否存在，<br/>存在则Del，最后通过Enum判断剩余元素，没有的话Del元信息

### List
指令 | IO | 性能 | 说明
---- | ---- | ---- | ----
LPUSH/RPUSH | S(n) S(1) |  |
LPOP/RPOP | G(1) D(1) S(1) |  | 
LTRIM | D(n) S(1) | | 
LINDEX | G(1) | | 
LRANGE | G(n) | | 
LLEN | 0 | | 

### ZSET
指令 | IO | 性能 | 说明
---- | ---- | ---- | ----
ZADD | G(n) D(n) S(n) S(n) S(1) |  | ZSET的实现最为复杂，需要用两个结构维护一个元素
ZCARD | 0 |  | 
ZRANK/ZREVRANK | E(1) | | 
ZRANGE/ZREVRANGE | E(1) | | 
ZRANGEBYSCORE<br/>ZREVRANGEBYSCORE | E(1) | | 
ZREM | G(n) D(n) D(n) S(1) |  | 
ZREMRANGEBYRANK<br/>ZREMRANGEBYSCORE | E(1) D(n) D(n) S(1) |  | 
ZINCRBY | D(1) S(2) S(1) |  | 
ZSCORE | G(1) |  | 










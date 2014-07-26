
#### KEY_NEXT
	
key_next用于扫描整个数据库，key顺序排序，使用上一次key_next返回的最后一个key，作为下一次key_next的[seek]可以实现扫描整个数据库，但要注意第二次key_next返回的第一个结果和上次最后一个结果相同，需要去重。

	keynext [seek] [count] [withtype] [withvalue]

	[seek] 要定位的起始key，或key前缀
	[count] 要返回的数量
	[withtype] 选填，返回key类型，string/hash/set/list/zset
	[withvalue] 选填，如果要扫描的对象是string，这里直接返回string内容，其它数据类型的返回值没有意义

实例：

	key_next '' 100 从开始扫描100条数据
	key_next 'user:100422:profile' 100 从指定的key开始向下扫描100条数据
	key_next '' 100 withtype 同时返回key类型，返回结果key,type,key,type,...
	key_next '' 100 withtype withvalue，同时返回key类型和key值，返回结果key,type,value,key,type,value,...



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










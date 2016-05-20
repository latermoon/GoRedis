package rocks

/*
基于RocksDB实现的Redis持久化层

1、key存储规则
为了提供keys、type等基本操作，每个存入的数据都会有这样的结构 +[key]type，用于表达key以及数据类型
比如一个set name latermoon，会在RocksDB里产生 +[name]string = latermoon 的数据
对于string以外的复杂结构，还会有另外的字段，比如 hash 会有以_h开头的key，list会有_l开头的key

2、RocksDB存储原则
因为整个设计都是为了海量存储的，所以所有支持的redis指令，都必须基于RocksDB实现，不能消耗内存
必要的时候，会牺牲掉一些redis特性，比如list结构需要lindex的话，就必须放弃lrem和linsert

同时会对使用场景进行一些取舍，比如zset要提供zcard的话，就需要每次操作后更新len，但增加的一次RocksDB操作会降低zadd性能
因此对于hash、set这种很少取count的数据，放弃hlen、scard的性能（但也可以提供1000以内的枚举统计）,来提高hset/sadd的性能

string
	+name,s = "latermoon"
hash
	+info,h = ""
	h[info]name = "latermoon"
	h[info]age = "27"
	h[info]sex = "M"
list
	+list,l = "0,3"
	l[list]0 = "a"
	l[list]1 = "b"
	l[list]2 = "c"
	l[list]3 = "d"
zset
	+user_rank,z = "3"
	z[user_rank]m#100422 = "-2"
	z[user_rank]m#100423 = "1"
	z[user_rank]m#300000 = "2"
	z[user_rank]s#-2#100422 = ""
	z[user_rank]s#1#100423 = ""
	z[user_rank]s#2#300000 = ""
*/

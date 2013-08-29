GoRedis
=======

### RedisServer Implemented by Go
#### 希望能解决的问题
1、实现Go版的Redis MOA，特定场合提供更好的性能
2、多机房数据同步，与海量数据备份
	一个GoRedis实例作为整个Redis集群（10台Redis）的从库，持久化到HBase，异地机房数据存储，并提供一定的查询性能
3、自定义双写策略

#### 开发中
		MongoStorage 实现Redis Get/Set 存储到MongoDB，之后应该有
		HBaseStorage 等待实现，慢速海量存储
		MySQLStorage 等待实现，慢速海量存储
		MultiSlaveOf 实现一个GoRedis作为n个Redis的从库，使用例子：作为10个Profile的从库，结合相应Storage实现海量冷存储

#### vi ~/.profile 

		export GOPATH=/User/lptmoon/Downloads/go/gopath/

#### Install:

		go get github.com/latermoon/GoRedis/goredis

#### Update:

		go get -u github.com/latermoon/GoRedis/goredis

#### Demo:

	server := goredis.NewRedisServer()

	// KeyValue
	kvCache := make(map[string]interface{})
	// Set操作的写锁
	chanSet := make(chan int, 1)

	server.On("GET", func(cmd *goredis.Command) (reply *goredis.Reply) {
		key := cmd.StringAtIndex(1)
		value := kvCache[key]
		reply = goredis.BulkReply(value)
		return
	})

	server.On("SET", func(cmd *goredis.Command) (reply *goredis.Reply) {
		key := cmd.StringAtIndex(1)
		value := cmd.StringAtIndex(2)
		chanSet <- 0
		kvCache[key] = value
		<-chanSet
		reply = goredis.StatusReply("OK")
		return
	})

	server.On("PING", func(cmd *goredis.Command) (reply *goredis.Reply) {
		reply = goredis.StatusReply("PONG")
		return
	})

	server.On("INFO", func(cmd *goredis.Command) (reply *goredis.Reply) {
		reply = goredis.BulkReply("GoRedis 0.1 by latermoon\n")
		return
	})

	// 开始监听端口
	server.Listen(":8002")

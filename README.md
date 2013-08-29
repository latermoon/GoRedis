GoRedis
=======

### RedisServer Implemented by Go
#### 说明
	1、本项目用于研究目的，累积出一套基于Redis的数据架构方案

#### 开发中
	MongoStorage 实现Redis Get/Set 存储到MongoDB
	MultiSlaveOf 实现一个GoRedis作为n个Redis的从库，使用例子：作为10个Profile的从库，结合相应Storage实现海量冷存储
	HBaseStorage 待实现，慢速海量存储
	MySQLStorage 待实现，慢速海量存储

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

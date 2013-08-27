package goredis

import ()

type SimpleRedisServer struct {
	// Keys
	OnDel    func(key string) (count int)     // Integer reply: The number of keys that were removed.
	OnExists func(key string) (exists int)    // Integer reply, 1 if the key exists.
	OnType   func(key string) (status string) // Status code reply: type of key, or none when key does not exist.
	// Strings
	OnGet    func(key string) (value interface{})       // Bulk reply: the value of key, or nil when key does not exist.
	OnSet    func(key string, value string) (err error) // Status code reply: OK if SET was executed correctly.
	OnMGet   func(keys ...string) (bulks []interface{}) // Multi-bulk reply: list of values at the specified keys.
	OnMSet   func(keyValues ...string) (status string)  // Status code reply: always OK since MSET can't fail.
	OnDecr   func(key string) (value int)               // Integer reply: the value of key after the decrement
	OnDecrBy func(key string, i int) (value int)        // Integer reply: the value of key after the decrement
	OnIncr   func(key string) (value int)               // Integer reply: the value of key after the increment
	OnIncrBy func(key string, i int) (value int)        // Integer reply: the value of key after the increment
	// Hashes
	OnHDel func(key string, fields ...string) (count int) // Integer reply: the number of fields that were removed from the hash, not including specified but non existing fields.
	// Lists
	OnLLen  func(key string) (length int)                   // Integer reply: the length of the list at key.
	OnLPop  func(key string) (value interface{})            // Bulk reply: the value of the first element, or nil when key does not exist.
	OnLPush func(key string, values ...string) (length int) // Integer reply: the length of the list after the push operations.
	OnRPop  func(key string) (value interface{})            // Bulk reply: the value of the last element, or nil when key does not exist.
	OnRPush func(key string, values ...string) (length int) // Integer reply: the length of the list after the push operation.
	// Sets
	// Sorted Sets
	// Pub/Sub
	// Transactions
	// Scripting
	// Connection
	OnAuth   func(password string) (status string) // Status code reply
	OnEcho   func(message string) (result string)  // Bulk reply
	OnPing   func() (result string)                // Status code reply "PONG"
	OnQuit   func() (status string)                // Status code reply: always OK.
	OnSelect func() (status string)                // Status code reply
	// Server
	OnInfo func(section string) (lines string) // Bulk reply: as a collection of text lines.
	OnSync func(s *Session)
	// 使用GoRedis实现
	innerServer *RedisServer
}

func NewSimpleRedisServer() (server *SimpleRedisServer) {
	server = &SimpleRedisServer{}

	server.innerServer = NewRedisServer()

	server.innerServer.On("GET", func(cmd *Command) (reply *Reply) {
		key := cmd.StringAtIndex(1)
		value := server.OnGet(key)
		reply = BulkReply(value)
		return
	})

	server.innerServer.On("SET", func(cmd *Command) (reply *Reply) {
		key := cmd.StringAtIndex(1)
		value := cmd.StringAtIndex(2)
		err := server.OnSet(key, value)
		if err != nil {
			reply = ErrorReply(err.Error())
		} else {
			reply = StatusReply("OK")
		}
		return
	})
	return
}

func (s *SimpleRedisServer) Listen(host string) {
	s.innerServer.Listen(host)
}

package goredis

import ()

type SimpleRedisServer struct {
	// Keys
	OnDEL    func(key string) (count int)     // Integer reply: The number of keys that were removed.
	OnEXISTS func(key string) (exists int)    // Integer reply, 1 if the key exists.
	OnTYPE   func(key string) (status string) // Status code reply: type of key, or none when key does not exist.
	// Strings
	OnGET    func(key string) (value interface{})       // Bulk reply: the value of key, or nil when key does not exist.
	OnSET    func(key string, value string) (err error) // Status code reply: OK if SET was executed correctly.
	OnMGET   func(keys ...string) (bulks []interface{}) // Multi-bulk reply: list of values at the specified keys.
	OnMSET   func(keyValues ...string) (status string)  // Status code reply: always OK since MSET can't fail.
	OnDECR   func(key string) (value int)               // Integer reply: the value of key after the decrement
	OnDECRBY func(key string, i int) (value int)        // Integer reply: the value of key after the decrement
	OnINCR   func(key string) (value int)               // Integer reply: the value of key after the increment
	OnINCRBY func(key string, i int) (value int)        // Integer reply: the value of key after the increment
	// Hashes
	OnHDEL func(key string, fields ...string) (count int) // Integer reply: the number of fields that were removed from the hash, not including specified but non existing fields.
	// Lists
	OnLLEN  func(key string) (length int)                   // Integer reply: the length of the list at key.
	OnLPOP  func(key string) (value interface{})            // Bulk reply: the value of the first element, or nil when key does not exist.
	OnLPUSH func(key string, values ...string) (length int) // Integer reply: the length of the list after the push operations.
	OnRPOP  func(key string) (value interface{})            // Bulk reply: the value of the last element, or nil when key does not exist.
	OnRPUSH func(key string, values ...string) (length int) // Integer reply: the length of the list after the push operation.
	// Sets
	OnSADD      func(key string, members ...string) (count int) // Integer reply: the number of elements that were added to the set
	OnSCARD     func(key string) (length int)                   // Integer reply: the cardinality (number of elements) of the set, or 0 if key does not exist.
	OnSISMEMBER func(key string) (exists int)                   // Integer reply, specifically: 1 if the element is a member of the set.
	OnSMEMBERS  func(key string) (bulks []interface{})          // Multi-bulk reply: all elements of the set.
	// Sorted Sets
	// Pub/Sub
	// Transactions
	OnMULTI func() (status string)       // Status code reply: always OK.
	OnEXEC  func() (bulks []interface{}) // Multi-bulk reply: each element being the reply to each of the commands
	// Scripting
	OnEVAL func(script string, numkeys int, keykeyvalval ...string) (reply *Reply) // http://redis.io/commands/eval
	// Connection
	OnAUTH   func(password string) (status string) // Status code reply
	OnECHO   func(message string) (result string)  // Bulk reply
	OnPING   func() (result string)                // Status code reply "PONG"
	OnQUIT   func() (status string)                // Status code reply: always OK.
	OnSELECT func() (status string)                // Status code reply
	// Server
	OnINFO func(section string) (lines string) // Bulk reply: as a collection of text lines.
	OnSYNC func(s *Session)
	// 使用GoRedis实现
	innerServer *RedisServer
}

func NewSimpleRedisServer() (server *SimpleRedisServer) {
	server = &SimpleRedisServer{}

	server.innerServer = NewRedisServer()

	server.innerServer.On("GET", func(cmd *Command) (reply *Reply) {
		key := cmd.StringAtIndex(1)
		value := server.OnGET(key)
		reply = BulkReply(value)
		return
	})

	server.innerServer.On("SET", func(cmd *Command) (reply *Reply) {
		key := cmd.StringAtIndex(1)
		value := cmd.StringAtIndex(2)
		err := server.OnSET(key, value)
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

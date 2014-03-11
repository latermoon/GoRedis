package goredis_server

// Session属性
const (
	S_STATUS = "status"
	// master
	REPL_WAIT      = "wait"
	REPL_SEND_BULK = "send_bulk"
	REPL_ONLINE    = "online"
	// slave
	REPL_CONNECT   = "connect"
	REPL_CONNECTED = "connected"
)

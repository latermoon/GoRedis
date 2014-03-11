package goredis_server

// Session属性
const (
	S_STATUS       = "status"
	REPL_WAIT      = "wait"
	REPL_SEND_BULK = "send_bulk" // master
	REPL_RECV_BULK = "recv_bulk" // slave
	REPL_ONLINE    = "online"
)

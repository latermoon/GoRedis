package goredis_server

// Session属性
const (
	S_STATUS       = "status"
	S_SLAVE_PORT   = "slaveport"
	S_HOST         = "host"
	S_LAST_COMMAND = "lastcmd"
	REPL_WAIT      = "wait"
	REPL_SEND_BULK = "send_bulk" // master
	REPL_RECV_BULK = "recv_bulk" // slave
	REPL_ONLINE    = "online"
)

// Command属性
const (
	C_SESSION = "session"
	C_ELAPSED = "elapsed"
)

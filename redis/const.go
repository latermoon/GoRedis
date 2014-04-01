package redis

/*
本类暂未使用

Datasource:
category,name,rw,min,max,oddeven
string,SET,1,

*/

// 指令集分类
const (
	C_Key         = "key"
	C_String      = "string"
	C_Hash        = "hash"
	C_List        = "list"
	C_Set         = "set"
	C_SortedSet   = "zset"
	C_PubSub      = "pubsub" // nouse
	C_Transaction = "trans"  // nouse
	C_Script      = "script" // nouse
	C_Connection  = "conn"
	C_Server      = "server"
	C_Unknown     = "unknown"
)

type CommandInfo struct {
	Name      string // SET/GET/...
	Category  string // key/string/set/hash/list/zset
	ReadWrite int    // 0 for read, 1 for write, -1 for undefined
	Min       int    // -1 for undefined
	Max       int    // -1 for undefined
	OddEven   int    // 0 for even, 1 for odd, -1 for undefined
}

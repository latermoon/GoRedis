package goredis

// ==============================
// 封装一个返回给客户端的Response
// 对于每种Redis响应，都有一个对象的构造函数
// ==============================
type Reply struct {
	Type  ReplyType
	Value interface{}
}

type ReplyType int

// 响应的种类
const (
	ReplyTypeStatus = iota
	ReplyTypeError
	ReplyTypeInteger
	ReplyTypeBulk
	ReplyTypeMultiBulks
)

func StatusReply(s string) (r *Reply) {
	r = &Reply{}
	r.Type = ReplyTypeStatus
	r.Value = s
	return
}

func ErrorReply(s string) (r *Reply) {
	r = &Reply{}
	r.Type = ReplyTypeError
	r.Value = s
	return
}

func IntegerReply(i int) (r *Reply) {
	r = &Reply{}
	r.Type = ReplyTypeInteger
	r.Value = i
	return
}

func BulkReply(bulk interface{}) (r *Reply) {
	r = &Reply{}
	r.Type = ReplyTypeBulk
	r.Value = bulk
	return
}

func MultiBulksReply(bulks []interface{}) (r *Reply) {
	r = &Reply{}
	r.Type = ReplyTypeMultiBulks
	r.Value = bulks
	return
}

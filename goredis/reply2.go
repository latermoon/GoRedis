package goredis

// ==============================
// 封装一个返回给客户端的Response
// 对于每种Redis响应，都有一个对应的构造函数
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

/**
 * @param status 绝大部分情况下status="OK"
 */
func StatusReply(status string) (r *Reply) {
	r = &Reply{}
	r.Type = ReplyTypeStatus
	r.Value = status
	return
}

/**
 * @param errmsg 返回具体的错误信息
 */
func ErrorReply(errmsg string) (r *Reply) {
	r = &Reply{}
	r.Type = ReplyTypeError
	r.Value = errmsg
	return
}

func IntegerReply(i int) (r *Reply) {
	r = &Reply{}
	r.Type = ReplyTypeInteger
	r.Value = i
	return
}

/**
 * @param bulk 数据可以是string或[]byte。对于string，会自动转换为[]byte发往客户端
 */
func BulkReply(bulk interface{}) (r *Reply) {
	r = &Reply{}
	r.Type = ReplyTypeBulk
	r.Value = bulk
	return
}

/**
 * @param bulks 数组元素可以是string, []byte, int, nil
 */
func MultiBulksReply(bulks []interface{}) (r *Reply) {
	r = &Reply{}
	r.Type = ReplyTypeMultiBulks
	r.Value = bulks
	return
}

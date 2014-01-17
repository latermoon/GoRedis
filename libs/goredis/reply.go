// Copyright (c) 2013, Latermoon <lptmoon@gmail.com>
// All rights reserved.
//
package goredis

import (
	"bytes"
	"fmt"
)

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
	ReplyTypeStatus ReplyType = iota
	ReplyTypeError
	ReplyTypeInteger
	ReplyTypeBulk
	ReplyTypeMultiBulks
)

var replyTypeDesc = map[ReplyType]string{ReplyTypeStatus: "StatusReply", ReplyTypeError: "ErrorReply", ReplyTypeInteger: "IntegerReply", ReplyTypeBulk: "BulkReply", ReplyTypeMultiBulks: "MultiBulksReply"}

func (r *Reply) String() string {
	buf := bytes.Buffer{}
	buf.WriteString("<")
	buf.WriteString(replyTypeDesc[r.Type])
	buf.WriteString(":>")
	return buf.String()
}

/**
 * 返回错误Reply或正确Reply（精简判断语句）
 * @param err 如果err != nil，返回ErrorReply，否则返回提供的正确Reply
 */
func ReplySwitch(err error, successReply *Reply) *Reply {
	if err != nil {
		return ErrorReply(err)
	} else {
		return successReply
	}
}

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
func ErrorReply(err interface{}) (r *Reply) {
	r = &Reply{}
	r.Type = ReplyTypeError
	switch err.(type) {
	case string:
		r.Value = err.(string)
	case error:
		r.Value = err.(error).Error()
	default:
		r.Value = fmt.Sprint(err)
	}
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

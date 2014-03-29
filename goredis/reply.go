// Copyright 2013 Latermoon. All rights reserved.

package goredis

import (
	"bytes"
	"fmt"
)

type ReplyType int

// 响应的种类
const (
	ReplyTypeStatus ReplyType = iota
	ReplyTypeError
	ReplyTypeInteger
	ReplyTypeBulk
	ReplyTypeMultiBulks
)

// 封装一个返回给客户端的Response
// 对于每种Redis响应，都有一个对应的构造函数
type Reply interface {
	Type() ReplyType
	Value() interface{}
}

var NOREPLY Reply = nil // 空回复

// 文本表示
var replyTypeDesc = map[ReplyType]string{
	ReplyTypeStatus:     "StatusReply",
	ReplyTypeError:      "ErrorReply",
	ReplyTypeInteger:    "IntegerReply",
	ReplyTypeBulk:       "BulkReply",
	ReplyTypeMultiBulks: "MultiBulksReply",
}

func NewReply(replyType ReplyType, value interface{}) (r Reply) {
	return &baseReply{
		replyType: replyType,
		value:     value,
	}
}

func StatusReply(status string) (r Reply) {
	return NewReply(ReplyTypeStatus, status)
}

func ErrorReply(err interface{}) (r Reply) {
	return NewReply(ReplyTypeError, fmt.Sprint(err))
}

func IntegerReply(i int) (r Reply) {
	return NewReply(ReplyTypeInteger, i)
}

// bulk 数据可以是string或[]byte。对于string，会自动转换为[]byte发往客户端
func BulkReply(bulk interface{}) (r Reply) {
	return NewReply(ReplyTypeBulk, bulk)
}

// bulks 数组元素可以是string, []byte, int, nil
func MultiBulksReply(bulks []interface{}) (r Reply) {
	return NewReply(ReplyTypeMultiBulks, bulks)
}

// ==============================
// 内部实现
type baseReply struct {
	Reply
	replyType ReplyType
	value     interface{}
}

func (r *baseReply) Type() ReplyType {
	return r.replyType
}

func (r *baseReply) Value() interface{} {
	return r.value
}

func (r *baseReply) String() string {
	buf := &bytes.Buffer{}
	buf.WriteString("<")
	buf.WriteString(replyTypeDesc[r.Type()])
	buf.WriteString(":")
	switch r.Value().(type) {
	case []byte:
		buf.WriteString(string(r.Value().([]byte)))
	default:
		buf.WriteString(fmt.Sprint(r.Value()))
	}
	buf.WriteString(">")
	return buf.String()
}

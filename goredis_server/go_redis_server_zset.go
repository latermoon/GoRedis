package goredis_server

// 本类编写仓促，重复代码较多，准备重构
// http://redis.io/commands#sorted_set

import (
	. "../goredis"
	"./libs/leveltool"
	"strings"
)

// ZADD key score member [score member ...]
// Add one or more members to a sorted set, or update its score if it already exists
func (server *GoRedisServer) OnZADD(cmd *Command) (reply *Reply) {
	key := cmd.StringAtIndex(1)
	scoreMembers := cmd.Args[2:]
	count := len(scoreMembers)
	if count%2 != 0 {
		return ErrorReply("Bad argument count")
	}

	zset := server.keyManager.zsetByKey(key)
	for i := 0; i < count; i += 2 {
		elem := leveltool.NewZSetElem(scoreMembers[i], scoreMembers[i+1])
		zset.Add(elem)
	}
	reply = IntegerReply(count / 2)
	return
}

func (server *GoRedisServer) OnZCARD(cmd *Command) (reply *Reply) {
	key := cmd.StringAtIndex(1)
	zset := server.keyManager.zsetByKey(key)
	reply = IntegerReply(zset.Count())
	return
}

// http://redis.io/commands/zrange
// ZRANGE key start stop [WITHSCORES]
// Return a range of members in a sorted set, by index
func (server *GoRedisServer) OnZRANGE(cmd *Command) (reply *Reply) {
	key := cmd.StringAtIndex(1)
	start, e1 := cmd.IntAtIndex(2)
	stop, e2 := cmd.IntAtIndex(3)
	if e1 != nil || e2 != nil {
		return ErrorReply("Bad start/stop")
	}
	// 输出score
	withScore := false
	if len(cmd.Args) >= 5 && strings.ToUpper(cmd.StringAtIndex(4)) == "WITHSCORES" {
		withScore = true
	}
	zset := server.keyManager.zsetByKey(key)
	// bulks
	bulks := make([]interface{}, 0, 10) // TODO 优化内存分配
	high2low := false
	elems := zset.RangeByIndex(high2low, start, stop)
	for _, elem := range elems {
		bulks = append(bulks, elem.Member)
		if withScore {
			bulks = append(bulks, elem.Score)
		}
	}
	reply = MultiBulksReply(bulks)
	return
}

// ZREVRANGE key start stop [WITHSCORES]
// Return a range of members in a sorted set, by index, with scores ordered from high to low
func (server *GoRedisServer) OnZREVRANGE(cmd *Command) (reply *Reply) {
	key := cmd.StringAtIndex(1)
	start, e1 := cmd.IntAtIndex(2)
	stop, e2 := cmd.IntAtIndex(3)
	if e1 != nil || e2 != nil {
		return ErrorReply("Bad start/stop")
	}
	// 输出score
	withScore := false
	if len(cmd.Args) >= 5 && strings.ToUpper(cmd.StringAtIndex(4)) == "WITHSCORES" {
		withScore = true
	}
	zset := server.keyManager.zsetByKey(key)
	// bulks
	bulks := make([]interface{}, 0, 10) // TODO 优化内存分配
	high2low := true
	elems := zset.RangeByIndex(high2low, start, stop)
	for _, elem := range elems {
		bulks = append(bulks, elem.Member)
		if withScore {
			bulks = append(bulks, elem.Score)
		}
	}
	reply = MultiBulksReply(bulks)
	return
}

// ZRANGEBYSCORE key min max [WITHSCORES] [LIMIT offset count]
// Return a range of members in a sorted set, by score
func (server *GoRedisServer) OnZRANGEBYSCORE(cmd *Command) (reply *Reply) {
	key := cmd.StringAtIndex(1)
	min, e1 := cmd.ArgAtIndex(2)
	max, e2 := cmd.ArgAtIndex(3)
	if e1 != nil || e2 != nil {
		return ErrorReply("Bad min/max")
	}
	// 输出score
	withScore := false
	if len(cmd.Args) >= 5 && strings.ToUpper(cmd.StringAtIndex(4)) == "WITHSCORES" {
		withScore = true
	}
	zset := server.keyManager.zsetByKey(key)
	// bulks
	bulks := make([]interface{}, 0, 10) // TODO 优化内存分配
	high2low := false
	elems := zset.RangeByScore(high2low, min, max, 0, -1)
	for _, elem := range elems {
		bulks = append(bulks, elem.Member)
		if withScore {
			bulks = append(bulks, elem.Score)
		}
	}
	reply = MultiBulksReply(bulks)
	return
}

// ZREVRANGEBYSCORE key max min [WITHSCORES] [LIMIT offset count]
// Return a range of members in a sorted set, by score, with scores ordered from high to low
func (server *GoRedisServer) OnZREVRANGEBYSCORE(cmd *Command) (reply *Reply) {
	key := cmd.StringAtIndex(1)
	min, e1 := cmd.ArgAtIndex(2)
	max, e2 := cmd.ArgAtIndex(3)
	if e1 != nil || e2 != nil {
		return ErrorReply("Bad min/max")
	}
	// 输出score
	withScore := false
	if len(cmd.Args) >= 5 && strings.ToUpper(cmd.StringAtIndex(4)) == "WITHSCORES" {
		withScore = true
	}
	zset := server.keyManager.zsetByKey(key)
	// bulks
	bulks := make([]interface{}, 0, 10) // TODO 优化内存分配
	high2low := true
	elems := zset.RangeByScore(high2low, min, max, 0, -1)
	for _, elem := range elems {
		bulks = append(bulks, elem.Member)
		if withScore {
			bulks = append(bulks, elem.Score)
		}
	}
	reply = MultiBulksReply(bulks)
	return
}

// ZREM key member [member ...]
// Remove one or more members from a sorted set
func (server *GoRedisServer) OnZREM(cmd *Command) (reply *Reply) {
	key := cmd.StringAtIndex(1)
	members := cmd.Args[2:]
	zset := server.keyManager.zsetByKey(key)
	n := zset.Remove(members...)
	reply = IntegerReply(n)
	return
}

// ZREMRANGEBYRANK key start stop
// Remove all members in a sorted set within the given indexes
func (server *GoRedisServer) OnZREMRANGEBYRANK(cmd *Command) (reply *Reply) {
	key := cmd.StringAtIndex(1)
	start, e1 := cmd.IntAtIndex(2)
	stop, e2 := cmd.IntAtIndex(3)
	if e1 != nil || e2 != nil {
		return ErrorReply("Bad start/stop")
	}
	zset := server.keyManager.zsetByKey(key)
	n := zset.RemoveByIndex(start, stop)
	reply = IntegerReply(n)
	return
}

// ZREMRANGEBYSCORE key min max
// Remove all members in a sorted set within the given scores
func (server *GoRedisServer) OnZREMRANGEBYSCORE(cmd *Command) (reply *Reply) {
	key := cmd.StringAtIndex(1)
	min, e1 := cmd.ArgAtIndex(2)
	max, e2 := cmd.ArgAtIndex(3)
	if e1 != nil || e2 != nil {
		return ErrorReply("Bad min/max")
	}
	zset := server.keyManager.zsetByKey(key)
	n := zset.RemoveByScore(min, max)
	reply = IntegerReply(n)
	return
}

func (server *GoRedisServer) OnZINCRBY(cmd *Command) (reply *Reply) {
	key := cmd.StringAtIndex(1)
	incrmemt, e1 := cmd.ArgAtIndex(2)
	member, e2 := cmd.ArgAtIndex(3)
	if e1 != nil || e2 != nil {
		return ErrorReply("Bad incrment/member")
	}
	zset := server.keyManager.zsetByKey(key)
	score := zset.IncrBy(incrmemt, member)
	reply = BulkReply(score)
	return
}

// ZSCORE key member
// Get the score associated with the given member in a sorted set
func (server *GoRedisServer) OnZSCORE(cmd *Command) (reply *Reply) {
	key := cmd.StringAtIndex(1)
	member, _ := cmd.ArgAtIndex(2)
	zset := server.keyManager.zsetByKey(key)
	score := zset.Score(member)
	if score == nil {
		return BulkReply(nil)
	}
	reply = BulkReply(score)
	return
}

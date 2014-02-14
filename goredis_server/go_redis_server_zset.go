package goredis_server

import (
	"./util"
	. "GoRedis/goredis"
	"strconv"
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
	args := make([][]byte, count)
	// format score
	for i := 0; i < count; i += 2 {
		scorefloat, err := strconv.ParseFloat(string(scoreMembers[i]), 64)
		if err != nil {
			return ErrorReply("bad score")
		}
		// replace score
		scoreInt := int64(scorefloat)
		args[i] = util.Int64ToBytes(scoreInt)
		args[i+1] = scoreMembers[i+1]
	}
	// add
	zset := server.levelRedis.GetSortedSet(key)
	n := zset.Add(args...)
	reply = IntegerReply(n)
	return
}

func (server *GoRedisServer) OnZCARD(cmd *Command) (reply *Reply) {
	key := cmd.StringAtIndex(1)
	zset := server.levelRedis.GetSortedSet(key)
	reply = IntegerReply(zset.Len())
	return
}

func (server *GoRedisServer) zrank(cmd *Command, high2low bool) (reply *Reply) {
	key := cmd.StringAtIndex(1)
	member, err := cmd.ArgAtIndex(2)
	if err != nil {
		return ErrorReply(err)
	}
	zset := server.levelRedis.GetSortedSet(key)
	idx := zset.Rank(high2low, member)
	if idx == -1 {
		return BulkReply(nil)
	} else {
		return IntegerReply(idx)
	}
}

func (server *GoRedisServer) OnZRANK(cmd *Command) (reply *Reply) {
	return server.zrank(cmd, false)
}

func (server *GoRedisServer) OnZREVRANK(cmd *Command) (reply *Reply) {
	return server.zrank(cmd, true)
}

func (server *GoRedisServer) rangeByIndex(cmd *Command, high2low bool) (reply *Reply) {
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
	zset := server.levelRedis.GetSortedSet(key)
	scoreMembers := zset.RangeByIndex(high2low, start, stop)
	count := len(scoreMembers)
	bulks := make([]interface{}, 0, count)
	for i := 0; i < count; i += 2 {
		bulks = append(bulks, scoreMembers[i+1])
		if withScore {
			scoreInt := util.BytesToInt64(scoreMembers[i])
			bulks = append(bulks, []byte(strconv.FormatInt(scoreInt, 10)))
		}
	}
	reply = MultiBulksReply(bulks)
	return
}

// http://redis.io/commands/zrange
// ZRANGE key start stop [WITHSCORES]
// Return a range of members in a sorted set, by index
func (server *GoRedisServer) OnZRANGE(cmd *Command) (reply *Reply) {
	return server.rangeByIndex(cmd, false)
}

// ZREVRANGE key start stop [WITHSCORES]
// Return a range of members in a sorted set, by index, with scores ordered from high to low
func (server *GoRedisServer) OnZREVRANGE(cmd *Command) (reply *Reply) {
	return server.rangeByIndex(cmd, true)
}

func (server *GoRedisServer) rangeByScore(cmd *Command, high2low bool) (reply *Reply) {
	key := cmd.StringAtIndex(1)
	min, e1 := cmd.Int64AtIndex(2)
	max, e2 := cmd.Int64AtIndex(3)
	if e1 != nil || e2 != nil {
		return ErrorReply("Bad min/max")
	}
	// 输出score
	withScore := false
	if len(cmd.Args) >= 5 && strings.ToUpper(cmd.StringAtIndex(4)) == "WITHSCORES" {
		withScore = true
	}
	zset := server.levelRedis.GetSortedSet(key)
	scoreMembers := zset.RangeByScore(high2low, util.Int64ToBytes(min), util.Int64ToBytes(max), 0, -1)
	count := len(scoreMembers)
	bulks := make([]interface{}, 0, count)
	for i := 0; i < count; i += 2 {
		bulks = append(bulks, scoreMembers[i+1])
		if withScore {
			scoreInt := util.BytesToInt64(scoreMembers[i])
			bulks = append(bulks, []byte(strconv.FormatInt(scoreInt, 10)))
		}
	}
	reply = MultiBulksReply(bulks)
	return
}

// ZRANGEBYSCORE key min max [WITHSCORES] [LIMIT offset count]
// Return a range of members in a sorted set, by score
func (server *GoRedisServer) OnZRANGEBYSCORE(cmd *Command) (reply *Reply) {
	return server.rangeByScore(cmd, false)
}

// ZREVRANGEBYSCORE key max min [WITHSCORES] [LIMIT offset count]
// Return a range of members in a sorted set, by score, with scores ordered from high to low
func (server *GoRedisServer) OnZREVRANGEBYSCORE(cmd *Command) (reply *Reply) {
	return server.rangeByScore(cmd, true)
}

// ZREM key member [member ...]
// Remove one or more members from a sorted set
func (server *GoRedisServer) OnZREM(cmd *Command) (reply *Reply) {
	key := cmd.StringAtIndex(1)
	members := cmd.Args[2:]
	zset := server.levelRedis.GetSortedSet(key)
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
	zset := server.levelRedis.GetSortedSet(key)
	n := zset.RemoveByIndex(start, stop)
	reply = IntegerReply(n)
	return
}

// ZREMRANGEBYSCORE key min max
// Remove all members in a sorted set within the given scores
func (server *GoRedisServer) OnZREMRANGEBYSCORE(cmd *Command) (reply *Reply) {
	key := cmd.StringAtIndex(1)
	min, e1 := cmd.Int64AtIndex(2)
	max, e2 := cmd.Int64AtIndex(3)
	if e1 != nil || e2 != nil {
		return ErrorReply("Bad min/max")
	}
	zset := server.levelRedis.GetSortedSet(key)
	n := zset.RemoveByScore(util.Int64ToBytes(min), util.Int64ToBytes(max))
	reply = IntegerReply(n)
	return
}

func (server *GoRedisServer) OnZINCRBY(cmd *Command) (reply *Reply) {
	key := cmd.StringAtIndex(1)
	incrmemt, e1 := cmd.Int64AtIndex(2)
	member, e2 := cmd.ArgAtIndex(3)
	if e1 != nil || e2 != nil {
		return ErrorReply("Bad incrment/member")
	}
	zset := server.levelRedis.GetSortedSet(key)
	score := zset.IncrBy(member, incrmemt)
	scoreInt := util.BytesToInt64(score)
	reply = BulkReply(strconv.FormatInt(scoreInt, 10))
	return
}

// ZSCORE key member
// Get the score associated with the given member in a sorted set
func (server *GoRedisServer) OnZSCORE(cmd *Command) (reply *Reply) {
	key, _ := cmd.ArgAtIndex(1)
	member, _ := cmd.ArgAtIndex(2)
	// zset := server.levelRedis.GetSortedSet(key)
	// score := zset.Score(member)
	score := server.levelRedis.Global().ZScore(key, member)
	if score == nil {
		return BulkReply(nil)
	}
	scoreInt := util.BytesToInt64(score)
	reply = BulkReply(strconv.FormatInt(scoreInt, 10))
	return
}

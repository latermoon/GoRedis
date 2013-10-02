package goredis_server

// http://redis.io/commands#sorted_set

import (
	. "../goredis"
	. "./storage"
	"strconv"
	"strings"
)

// 获取SortedSet，不存在则自动创建
func (server *GoRedisServer) sortedSetByKey(key string) (sse *SortedSetEntry, err error) {
	entry := server.datasource.Get(key)
	if entry != nil && entry.Type() != EntryTypeSortedSet {
		err = WrongKindError
		return
	}
	if entry == nil {
		entry = NewSortedSetEntry()
		server.datasource.Set(key, entry)
	}
	sse = entry.(*SortedSetEntry)
	return
}

// ZADD key score member [score member ...]
// Add one or more members to a sorted set, or update its score if it already exists
func (server *GoRedisServer) OnZADD(cmd *Command) (reply *Reply) {
	key := cmd.StringAtIndex(1)
	scoreMembers := cmd.StringArgs()[2:]
	count := len(scoreMembers)
	if count%2 != 0 {
		return ErrorReply("Bad argument count")
	}
	entry, err := server.sortedSetByKey(key)
	if err != nil {
		return ErrorReply(err)
	}
	for i := 0; i < count; i += 2 {
		score, e1 := strconv.Atoi(scoreMembers[i])
		if e1 != nil {
			return ErrorReply(e1)
		}
		member := scoreMembers[i+1]
		entry.SkipList().Set(score, member)
	}
	// The number of elements added to the sorted sets
	reply = IntegerReply(count / 2)
	return
}

func (server *GoRedisServer) OnZCARD(cmd *Command) (reply *Reply) {
	key := cmd.StringAtIndex(1)
	entry, err := server.sortedSetByKey(key)
	if err != nil {
		return ErrorReply(err)
	}
	count := entry.SkipList().Len()
	reply = IntegerReply(count)
	return
}

func (server *GoRedisServer) OnZCOUNT(cmd *Command) (reply *Reply) {
	return
}

// http://redis.io/commands/zrange
// ZRANGE key start stop [WITHSCORES]
// Return a range of members in a sorted set, by index
func (server *GoRedisServer) OnZRANGE(cmd *Command) (reply *Reply) {
	key := cmd.StringAtIndex(1)
	entry, err := server.sortedSetByKey(key)
	if err != nil {
		return ErrorReply(err)
	}
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
	i := 0
	bulks := make([]interface{}, 0, 100) // TODO 优化内存分配
	for iter := entry.SkipList().Iterator(); iter.Next(); {
		if i >= start && (stop == -1 || i <= stop) {
			bulks = append(bulks, iter.Value())
			if withScore {
				bulks = append(bulks, iter.Key())
			}
		}
		i++
	}
	reply = MultiBulksReply(bulks)
	return
}

// ZRANGEBYSCORE key min max [WITHSCORES] [LIMIT offset count]
// Return a range of members in a sorted set, by score
func (server *GoRedisServer) OnZRANGEBYSCORE(cmd *Command) (reply *Reply) {
	key := cmd.StringAtIndex(1)
	entry, err := server.sortedSetByKey(key)
	if err != nil {
		return ErrorReply(err)
	}
	min, e1 := cmd.IntAtIndex(2)
	max, e2 := cmd.IntAtIndex(3)
	if e1 != nil || e2 != nil {
		return ErrorReply("Bad min/max")
	}
	// 输出score
	withScore := false
	if len(cmd.Args) >= 5 && strings.ToUpper(cmd.StringAtIndex(4)) == "WITHSCORES" {
		withScore = true
	}
	//iter := entry.SkipList().Iterator()
	bulks := make([]interface{}, 0, 100) // TODO 优化内存分配
	iter := entry.SkipList().Range(min, max)
	for iter.Next() {
		bulks = append(bulks, iter.Value())
		if withScore {
			bulks = append(bulks, iter.Key())
		}
	}
	reply = MultiBulksReply(bulks)
	return
}

func (server *GoRedisServer) OnZREM(cmd *Command) (reply *Reply) {
	// key := cmd.StringAtIndex(1)
	// entry, err := server.sortedSetByKey(key)
	// if err != nil {
	// 	return ErrorReply(err)
	// }
	// members := cmd.StringArgs()[2:]

	return
}

func (server *GoRedisServer) OnZREMRANGEBYSCORE(cmd *Command) (reply *Reply) {
	return
}

func (server *GoRedisServer) OnZREVRANGE(cmd *Command) (reply *Reply) {
	return
}

// ZREVRANGEBYSCORE key max min [WITHSCORES] [LIMIT offset count]
// Return a range of members in a sorted set, by score, with scores ordered from high to low
func (server *GoRedisServer) OnZREVRANGEBYSCORE(cmd *Command) (reply *Reply) {
	return
}

// ZSCORE key member
// Get the score associated with the given member in a sorted set
func (server *GoRedisServer) OnZSCORE(cmd *Command) (reply *Reply) {
	return
}

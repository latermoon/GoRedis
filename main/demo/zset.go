package main

import (
	"../../goredis_server/libs/sortedset"
	"fmt"
)

func main() {
	zset := sortedset.NewSortedSet()
	zset.Add("A", 100)
	zset.Add("C", 11)
	zset.Add("B", 2)
	zset.Add("C", 11)
	zset.Add("C", 11)
	zset.Add("D", 103.1)
	zset.Add("C", 11)

	iter1 := zset.RangeByScore(0, 100)
	printIterator(iter1)

	//printSkipList(zset)
}

func printIterator(iter sortedset.Iterator) {
	for iter.Next() {
		fmt.Println(iter.Key(), iter.Value())
	}
}

func printSkipList(s *sortedset.SortedSet) {
	fmt.Println("==============>")
	for iter := s.Iterator(); iter.Next(); {
		fmt.Println(iter.Key(), iter.Value())
	}
	fmt.Println("<==============")
}

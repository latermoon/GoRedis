package leveltool

import (
	"bytes"
	"github.com/syndtr/goleveldb/leveldb/iterator"
)

// 通用的key前缀枚举函数，设置*quit=true时立即退出枚举
// @param direction 枚举方向，"prev"从大到小，"next"表示从小到大
func PrefixEnumerate(iter iterator.Iterator, prefix []byte, fn func(i int, iter iterator.Iterator, quit *bool), direction string) {
	iter.Seek(prefix)
	i := 0
	if bytes.HasPrefix(iter.Key(), prefix) {
		quit := false
		fn(i, iter, &quit)
		if quit {
			return
		}
	}
	searchPrev := direction == "prev"
	for {
		hasMore := false
		if !searchPrev {
			hasMore = iter.Next()
		} else {
			hasMore = iter.Prev()
		}
		if !hasMore {
			break
		}
		if bytes.HasPrefix(iter.Key(), prefix) {
			i++
			quit := false
			fn(i, iter, &quit)
			if quit {
				return
			}
		} else {
			break
		}
	}
}

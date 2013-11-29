package leveltool

import (
	"bytes"
	"github.com/syndtr/goleveldb/leveldb/iterator"
	"strings"
)

func joinStringBytes(s ...string) []byte {
	return []byte(strings.Join(s, ""))
}

func joinBytes(b ...[]byte) []byte {
	return bytes.Join(b, nil)
}

// 前缀扫描，设置*quit=true时退出枚举
// @param direction 枚举方向，"prev"从大到小，默认值"next"表示从小到大
func PrefixEnumerate(iter iterator.Iterator, prefix []byte, fn func(i int, iter iterator.Iterator, quit *bool), direction string) {
	searchPrev := direction == "prev"
	var seekkey []byte
	if searchPrev {
		// 定位到下一个大于当前prefix的虚拟key
		seekkey = append(prefix, 254)
	} else {
		seekkey = prefix
	}
	found := iter.Seek(seekkey)
	i := -1
	if found && bytes.HasPrefix(iter.Key(), prefix) {
		i++
		quit := false
		fn(i, iter, &quit)
		if quit {
			return
		}
	}

	for {
		hasMore := false
		if searchPrev {
			hasMore = iter.Prev()
		} else {
			hasMore = iter.Next()
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

// 范围枚举，不要求前缀相同
func RangeEnumerate(iter iterator.Iterator, min, max []byte, fn func(i int, iter iterator.Iterator, quit *bool), high2low bool) {
	var found bool
	if high2low {
		max = append(max, 254)
		found = iter.Seek(max)
	} else {
		found = iter.Seek(min)
	}
	i := -1
	if found && bytes.Compare(iter.Key(), min) > 0 && bytes.Compare(iter.Key(), max) < 0 {
		i++
		quit := false
		fn(i, iter, &quit)
		if quit {
			return
		}
	}

	for {
		found = false
		if high2low {
			found = iter.Prev()
			if !found || bytes.Compare(iter.Key(), min) < 0 {
				break
			}
		} else {
			found = iter.Next()
			if !found || bytes.Compare(iter.Key(), max) > 0 {
				break
			}
		}
		i++
		quit := false
		fn(i, iter, &quit)
		if quit {
			return
		}
	}
	return
}

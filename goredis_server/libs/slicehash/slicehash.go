package slicehash

// 当原始个数少于n个时，内部使用slice存储，大于n个时，使用map
type SliceHash struct {
	slice []interface{}
	table map[interface{}]interface{}
}

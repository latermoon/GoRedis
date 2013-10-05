package goredis_server

// 通过leveldb实现的aof
type SlaveAof struct {
	prefix    string
	currentId int
}

func NewSlaveAof(prefix string) (aof *SlaveAof) {
	aof = &SlaveAof{}
	aof.prefix = prefix
	return
}

func (aof *SlaveAof) Append(entry []byte) (uid int) {
	return
}

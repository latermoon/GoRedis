package levelredis

// 这里放快捷指令，直接操作内部数据，减少初始化成本
type global struct {
	redis *LevelRedis
}

func newGlobal(redis *LevelRedis) (g *global) {
	g = &global{}
	g.redis = redis
	return
}

func (g *global) ZScore(key, member []byte) (score []byte, err error) {
	score, err = g.redis.RawGet(zmemberKey(key, member))
	return
}

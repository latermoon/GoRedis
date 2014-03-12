package shardredis

/*
shardredis.Load(fd)
cluster := shardredis.Get("redis-profile-b")
rd := cluster.Get("100422")
reply, err := rd.Do("SET", "name", "latermoon")
rd.Close()

*/
type ShardRedis struct {
}

func Load() {

}

func Get(name string) (c *Cluster) {
	return
}

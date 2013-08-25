package goredis

type RedisServerEx struct {
	SimpleRedisServer
	OnGet func(s *Session, key string) (value interface{}, err error)
	OnSet func(s *Session, key string, value string) (err error)
	OnDel func(s *Session, key string) (err error)
}

func NewRedisServerEx() (server *RedisServerEx) {
	server = &RedisServerEx{}
	server.SimpleRedisServer.Init()
	server.On("GET", func(s *Session, cmd *Command) (err error) {
		key, _ := cmd.StringAtIndex(1)
		value, err := server.OnGet(s, key)
		if err != nil {
			s.ReplyError(err.Error())
		} else {
			s.ReplyBulk(value)
		}
		return
	})

	server.On("SET", func(s *Session, cmd *Command) (err error) {
		key, _ := cmd.StringAtIndex(1)
		value, _ := cmd.StringAtIndex(1)
		err = server.OnSet(s, key, value)
		if err != nil {
			s.ReplyError(err.Error())
		} else {
			s.ReplyStatus("OK")
		}
		return
	})
	return
}

func (r *RedisServerEx) Listen(host string) {
	r.SimpleRedisServer.Listen(host)
}

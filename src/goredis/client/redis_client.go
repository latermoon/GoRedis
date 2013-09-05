package client

import (
	"net"
)

type RedisClient struct {
	conn net.Conn
}

func NewRedisClient(conn net.Conn) (client *RedisClient) {
	client = *RedisClient{}
	client.conn = conn
	return
}

func (r *RedisClient) Do(cmd string, args ...string) (reply interface{}, err error) {
	return
}

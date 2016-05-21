package main

import (
	"flag"
	"github.com/latermoon/GoRedis/redis"
	"github.com/latermoon/GoRedis/rocks"
	"github.com/latermoon/GoRedis/server"
	"github.com/tecbot/gorocksdb"
	"log"
	"net"
)

func init() {
	flag.StringVar(&address, "bind address", ":6380", "Bind address")
}

func main() {
	flag.Parse()
	log.Println("server start ...")

	// new rocksdb
	db := newRocksDB("/tmp/rocks_6380")

	// new GoRedisServer handler
	handler := server.New(db)

	// register command handler
	redis.Register(handler)

	// Serve
	lis, err := net.Listen("tcp", address)
	if err != nil {
		panic(err)
	}
	redis.Serve(lis)
}

func newRocksDB(dir string) *rocks.DB {
	opts := gorocksdb.NewDefaultOptions()
	opts.SetCreateIfMissing(true)
	rdb, err := gorocksdb.OpenDb(opts, dir)
	if err != nil {
		panic(err)
	}
	return rocks.New(rdb)
}

var (
	address string
)

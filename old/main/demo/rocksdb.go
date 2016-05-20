package main

import (
	"GoRedis/libs/stdlog"
	levigo "github.com/bsm/go-rocksdb"
)

func main() {
	opts := levigo.NewOptions()
	opts.SetCache(levigo.NewLRUCache(128 * 1024 * 1024))
	opts.SetCompression(levigo.SnappyCompression)
	opts.SetBlockSize(32 * 1024)
	opts.SetWriteBufferSize(128 * 1024 * 1024)
	opts.SetMaxOpenFiles(100000)
	opts.SetCreateIfMissing(true)

	db, e1 := levigo.Open("/tmp/rocksdb0", opts)
	if e1 != nil {
		panic(e1)
	}
	stdlog.Println(db)

	batch := levigo.NewWriteBatch()
	batch.Put([]byte("name"), []byte("later"))
	wo := levigo.NewWriteOptions()
	db.Write(wo, batch)

	ro := levigo.NewReadOptions()
	value, err := db.Get(ro, []byte("name"))
	stdlog.Println(string(value), err)

	db.Close()
}

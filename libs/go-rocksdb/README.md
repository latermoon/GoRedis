# go-rocksdb

go-rocksdb is a Go wrapper for [RocksDB](http://rocksdb.org).

This library is an almost exact copy of [levigo](http://github.com/jmhodges/levigo) only built for RocksDB, not LevelDB.

## Building

You'll need the shared library build of
[RocksDB](https://github.com/facebook/rocksdb) installed on your machine. The
current RocksDB will build it by default.

Now, if you build RocksDB and put the shared library and headers in one of the
standard places for your OS, you'll be able to simply run:

    go get github.com/bsm/go-rocksdb

But, suppose you put the shared RocksDB library somewhere weird like
/path/to/lib and the headers were installed in /path/to/include. To install
go-rocksdb remotely, you'll run:

    CGO_CFLAGS="-I/path/to/rocksdb/include" CGO_LDFLAGS="-L/path/to/rocksdb/lib" go get github.com/bsm/go-rocksdb

and there you go.

In order to build with snappy, you'll have to explicitly add "-lsnappy" to the
`CGO_LDFLAGS`. Supposing that both snappy and rocksdb are in weird places,
you'll run something like:

    CGO_CFLAGS="-I/path/to/rocksdb/include -I/path/to/snappy/include"
    CGO_LDFLAGS="-L/path/to/rocksdb/lib -L/path/to/snappy/lib -lsnappy" go get github.com/bsm/go-rocksdb

(and make sure the -lsnappy is after the snappy library path!).

Of course, these same rules apply when doing `go build`, as well.


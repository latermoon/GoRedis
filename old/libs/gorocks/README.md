# gorocks

gorocks is a Go wrapper for RocksDB.

It is based on levigo by Jeff Hodges.

The API has been godoc'ed and [is available on the
web](http://godoc.org/github.com/alberts/gorocks).

## Building

    CGO_CFLAGS="-I/path/to/rocksdb/include" CGO_LDFLAGS="-L/path/to/rocksdb" go get github.com/alberts/gorocks





latermoon:redis $ go install "GoRedis/libs/gorocks"
# GoRedis/libs/gorocks
38: error: use of undeclared identifier 'rocksdb_options_enable_statistics'
38: error: use of undeclared identifier 'rocksdb_options_set_stats_dump_period_sec'
38: error: use of undeclared identifier 'rocksdb_options_set_max_bytes_for_level_base'
38: error: use of undeclared identifier 'rocksdb_options_set_min_level_to_compress'
38: error: use of undeclared identifier 'rocksdb_readoptions_set_tailing'; did you mean 'rocksdb_readoptions_set_snapshot'?
# install GoRedis into $GOPATH

echo "install goredis libs ..."

go install "GoRedis/goredis"
go install "GoRedis/libs/counter"
go install "GoRedis/libs/funcpool"
go install "GoRedis/libs/geo"
go install "GoRedis/libs/iotool"
go install "GoRedis/libs/jsonconf"
go install "GoRedis/libs/levelredis"
go install "GoRedis/libs/lrucache"
go install "GoRedis/libs/ratelimit"
go install "GoRedis/libs/redigo/redis"
go install "GoRedis/libs/redigo/redisx"
go install "GoRedis/libs/redis_tool"
go install "GoRedis/libs/shardredis"
go install "GoRedis/libs/rdb"
go install "GoRedis/libs/rdb/crc64"
go install "GoRedis/libs/stat"
go install "GoRedis/libs/stdlog"
go install "GoRedis/libs/uuid"
go install "GoRedis/goredis_server"
go install "GoRedis/goredis_proxy"

# 编译gorocks需要编译好rocksdb，并且配置环境变量
# export CGO_CFLAGS="-I/home/download/rocksdb/include/"
# export CGO_LDFLAGS="-L/home/download/rocksdb/ -lsnappy -lgflags -lz -lbz2 -lrt"
# debug
# export CGO_CFLAGS="-I/Users/latermoon/Downloads/dev/rocksdb/include/"
# export CGO_LDFLAGS="-L/Users/latermoon/Downloads/dev/rocksdb/ -lsnappy -lgflags -lz -lbz2"
go install "GoRedis/libs/gorocks"

echo "ok"


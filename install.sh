# install GoRedis into $GOPATH

echo "install goredis libs ..."

go install "GoRedis/goredis"
go install "GoRedis/libs/iotool"
go install "GoRedis/libs/levelredis"
go install "GoRedis/libs/lrucache"
go install "GoRedis/libs/redis_tool"
go install "GoRedis/libs/safelist"
go install "GoRedis/libs/statlog"
go install "GoRedis/libs/stdlog"
go install "GoRedis/libs/uuid"
go install "GoRedis/libs/queueprocess"
go install "GoRedis/libs/rdb"
go install "GoRedis/libs/rdb/crc64"
go install "GoRedis/libs/geo"

echo "ok"


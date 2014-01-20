
echo "build goredis-server bin ..."

target=/home/server/goredis/goredis-server
go build -o $target main/goredis-server.go

echo "ok"

echo "version:"
chmod +x $target
$target -v
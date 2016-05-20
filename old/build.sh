
echo "build goredis-server bin ..."

target=/home/server/goredis/goredis-server$1
go build -o $target main/goredis-server.go

echo "ok"

echo "version:"
chmod +x $target
ver=`$target -v | awk '{print $2}'`
cp $target $target'-'$ver
echo 'goredis-server' $ver

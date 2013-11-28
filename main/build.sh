
dt=`date +%Y.%m.%d`
target=goredis-server-$dt
go build -o $target goredis-server.go

cp $target /home/server/goredis/goredis-server
mv $target /home/server/goredis/$target

echo finish
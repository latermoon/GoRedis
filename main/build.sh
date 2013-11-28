

dt=`date +%Y.%m.%d`
target=goredis-server-$dt
go build -o $target goredis-server.go

mkdir /home/server/goredis/ -p
rm -f /home/server/goredis/goredis-server 
rm -f /home/server/goredis/$target

cp $target /home/server/goredis/goredis-server 
mv $target /home/server/goredis/$target

echo finish
### slaveof-proxy使用

slaveof-proxy为两个redis建立自定义的主从同步，包含限速、断线重试等。

临时安装：

	cd /home/server/goredis/bin/
	wget http://ssd002:8801/slaveof-proxy -O slaveof-proxy -q
	chmod +x slaveof-proxy

运行：

	./slaveof-proxy -src master:port -dest slave:port -pullrate 240 -buffer 100

**-pullrate** 表示拉取主库RDB时的限速，单位是Mbits/s，比如-pullrate 240表示限速30MB/s，一般内网千兆网卡-pullrate最大设为600，这里最小为100

**-pushrate** 表示推送到从库的限速，目前没开发

**-buffer** 表示推送从库的缓冲区，单位万条，默认缓存100万条指令，如果从库网络中断，只要主库的buffer未满，重启从库后可以继续接收，实现跨机房可靠传输。

结合上面两个限速，slaveof-proxy进程可以运行在主库机房，或者从库机房，但为了实现专线网络瞬间中断，最适合是放到主库机房。
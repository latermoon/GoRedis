### SLAVEOF-PROXY使用

slaveof-proxy为两个redis建立自定义的主从同步，包含限速、断线重试等。

临时安装：

	cd /home/server/goredis/bin/
	wget http://ssd002:8801/slaveof-proxy -O slaveof-proxy -q
	chmod +x slaveof-proxy

运行：

	./slaveof-proxy -src master:port -dest slave:port -pullrate 400 -pushrate 400

**-pullrate** 表示拉取主库RDB时的限速，单位是Mbits/s，比如-pullrate 400表示限速400Mb(50MB/s)，一般内网千兆网卡-pullrate最大设为600，默认值400

**-pushrate** 表示推送到从库的限速，单位是Mbits/s，默认值400

**-buffer** 表示从库推送的缓冲区，默认值100，单位万条，表示可以100万条命令，如果从库网络中断，只要主库的buffer未满，从库启动后可以继续接收，实现跨机房可靠传输。

结合上面两个限速，slaveof-proxy进程可以运行在主库机房，或者从库机房，但为了实现专线网络瞬间中断，最适合是放到主库相同的服务器上。

	如果在主库服务器上运行，可以取消拉取限速，-pullrate 1000
	如果在从库服务器上运行，可以取消推送限速，-pushrate 1000
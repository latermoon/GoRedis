
### 安装GoRedis
	wget ssd002:8801/install_goredis.sh -O install_goredis.sh
	sh install_goredis.sh
### 运行GoRedis
	export LD_LIBRARY_PATH=$LD_LIBRARY_PATH:.:/home/server/goredis/bin
	./goredis-server

### Supervisor配置
	[program:goredis]
	command=/home/server/goredis/bin/goredis-server -procs 6 -p %(process_num)04d
	process_name=%(program_name)s-%(process_num)04d
	numprocs=5
	numprocs_start=18400
	autostart=false
	autorestart=true
	stdout_logfile=NONE
	stderr_logfile=/home/server/goredis/log/goredis_%(process_num)04d-stderr.log
	stdout_logfile_maxbytes=500MB
	stdout_logfile_backups=50
	stdout_capture_maxbytes=1MB
	stdout_events_enabled=false
	loglevel=info
	priority=1100

### install_goredis.sh
	echo 'downloading ...'
	mkdir /home/server/goredis/ -p
	cd /home/server/goredis/
	mkdir bin etc log -p
	cd bin
	wget ssd002:8801/goredis-server -O goredis-server -q
	chmod +x goredis-server
	wget http://ssd002:8801/libgflags.so.2 -O libgflags.so.2 -q
	wget http://ssd002:8801/librocksdb.so -O librocksdb.so -q
	wget http://ssd002:8801/libsnappy.so.1 -O libsnappy.so.1 -q
	export LD_LIBRARY_PATH=$LD_LIBRARY_PATH:.
	./goredis-server -v

### 杂项
	export LD_LIBRARY_PATH=$LD_LIBRARY_PATH:.:/usr/local/lib:/usr/lib:/usr/lib64
	go tool pprof goredis-server /tmp/goredis_1602/mem.prof

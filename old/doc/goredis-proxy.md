### GOREDIS-PROXY使用

GoRedis-Proxy为Redis主从实例提供访问端口，并可以动态配置主从和读写模式。适用于实现Redis高可用、数据统计访问，以及数据迁移。

	./goredis-proxy -master localhost:1602 -slave localhost:1603 -port 1602 -mode rrw

#### 管理指令

	> info
	> config master localhost:1602
	> config slave localhost:1603
	> config mode r/rr/rw/rrw

#### CONFIG

	> config master localhost:1602 //去掉原来的主库，设置新主库
	> config slave localhost:1603 //去掉原来的从库，设置新从库

#### CONFIG MODE
	
	> config mode rrw
	设置主从读写模式，mode包含rr表示主从均提供读操作，包含w表示主库允许写入，任何mode下，从库读出错都会自动读主库。
	mode=r, 从库提供读，写操作返回错误
	mode=rr, 主从均提供读，写操作返回错误
	mode=rw, 主库提供写，从库提供读
	mode=rrw，(default)主库提供写，主从均提供读

#### INFO

	> info
	# Server
	proxy_version:1.0.4
	mode:rrw
	# Master
	master:127.0.0.1,6379,OK
	master_uptime_in_seconds:44
	master_total_commands:34140
	master_ops_per_sec:13484
	# Slave0
	slave0:127.0.0.1,6389,OK
	slave0_uptime_in_seconds:44
	slave0_total_commands:34432
	slave0_ops_per_sec:13559

#### 维护说明
	
	

	
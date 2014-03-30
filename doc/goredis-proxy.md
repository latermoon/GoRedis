### goredis-proxy使用

	./goredis-proxy -mode rw/rrw -master localhost:1602 -slave localhost:1603


#### 管理指令

	> info
	> config master localhost:1602
	> config slave localhost:1603
	> config mode rw/rrw
### GoRedis主从复制

GoRedis和GoRedis之间可以实现一主多从或一从多主，同时同步过程使用带版本号的AOF保障断线后增量同步。

GoRedis可以作为一个或多个Redis的从库，但暂时不可以作为Redis的主库。


### 配置方式

目前只支持


#### GoRedis与GoRedis

##### 同步机制

同步指令和Redis一致，从库输入SLAVEOF host port，主库收到来自从库的SYNC指令后先把本地的数据快照发送给从库，之后把新增指令实时传送给从库。

	S: SYNC [UID] [SEQ]			// 从库向主库发送自身的UID，以及希望同步的起始SEQ，如果之前没同步过，SEQ=-1
	M: SYNC_RAW_BEG				// 主库告知开始传输快照数据
	M: SYNC_RAW [KEY] [VALUE]	// 主库发送快照数据，这里会持续很长时间
	M: ...
	M: SYNC_SEQ_BEG				// 快照结束后，告知从库开始接收带版本号的实时数据
	M: SYNC_SEQ [SEQ]			// 对于每个要发送的[CMD]，会先发送SEQ
	M: [CMD]					// 每个实时同步的[CMD]前必定带有SEQ
	M: ...						

从库需要不断更新最后收到的SEQ，断线后向主库SYNC [UID] [Last SEQ]，此时不需要接收快照，会直接进入SYNC_SEQ_BEG。

##### 同步机制v2

支持双主同步，从库同步主库快照后，订阅主库的SEQ更新，主库也订阅从库的SEQ更新，实现双主同步。

	SLAVEOF
	S: SYNC UID [UID] PORT [PORT] SNAP [1/0]
	M: SYNC_RAW_START
	M: SYNC_RAW [KEY] [VALUE]
	M: SYNC_RAW_END [COUNT] [SYNC-SEQ] [LAST-SEQ]

	SYNCOF
	S: SYNC SEQ [-1/...]
	M: SYNC_SEQ_START
	M: SYNC_SEQ [SEQ]
	M: [CMD]
	M: ...



##### 同步性能

在200G的SLC SSD下，SYNC_RAW性能10w/s，但要注意SYNC_RAW是rocksdb原始数据，并非面向外部的Redis数据，一个带4个元素的hash，会有5条RAW数据，在内网同步速率大约100Mb/s。










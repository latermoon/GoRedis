package goredis_server

/*
自定义geo指令集，用于实现附近位置搜索
geo_insert mapname lat lng key time [key1 value1 ...] <StatusReply: OK>
geo_nearby mapname lat lng start end
	1) 100422
	2) 300000
geo_del mapname value

###

geo:user_f15m:count = 100
geo:user_f15m:hash:9rj5jgn5842p:value:0 = [key ]
geo:user_f15m:hash:wx4g4hcvzfxg:value:1 = 300000,1378002003135
...
geo:user_f15m:key:100422:hash = 9rj5jgn5842p
geo:user_f15m:key:300000:hash = wx4g4hcvzfxg
...
geo:user_f15m:time:1378000907596:key:0 = 100422
geo:user_f15m:time:1378000907596:key:1 = 100428
geo:user_f15m:time:1378002003135:key:0 = 300000

*/

package main

import (
	//"encoding/json"
	"fmt"
	"github.com/garyburd/redigo/redis"
	"time"
)

var profileJson = "{\"_id\":{\"$id\":\"501f8365ccd569a138000000\"},\"age\":28,\"background\":\"447DA905-6927-F3BF-E64F-BF4CBC7911C2\",\"bind_email\":\"YES\",\"birthday\":{\"sec\":491932800,\"usec\":0},\"company\":\"\\u964c\\u964c\",\"constellation\":\"\\u72ee\\u5b50\\u5ea7\",\"device\":{\"client\":\"ios\",\"uid\":\"d2d7b90297b5f9522c873b7166c30864\",\"version\":96},\"email\":\"latermoon@qq.com\",\"hangout\":\"\",\"industry\":\"I1\",\"interest\":\"\\u521b\\u9020\\u4e00\\u4e9b\\u5b9e\\u7528\\u6709\\u8da3\\u7684\\u4e8b\\u7269\\n\\u5de5\\u4f5c\\uff0c\\u5bb6\\u5ead\\uff0c\\u7231\\u597d\",\"invitecode\":\"c7f1512c\",\"job\":\"\\u964c\\u964c\",\"momoid\":\"300000\",\"name\":\"Latermoon\\u2615\",\"phone\":{\"countrycode\":\"+86\",\"phonenumber\":\"18611308844\",\"bind_time\":{\"sec\":1361421527,\"usec\":335000},\"type\":2},\"photos\":[\"E180143B-11BC-C7F9-CC55-D3C750DD5F14\",\"94686DDD-E788-577E-EB5F-4A9360F6B075\",\"01AFA030-9B8D-95F0-DE6A-0828F59B14F9\",\"0C143347-4850-9E17-8D39-5678C26F77AA\",\"29B3EA51-1659-1266-4CBE-E3D5E6B20168\",\"9161AFE3-5C77-F7E8-89D2-38166EB4B362\",\"8DFB490C-1B2C-39D9-AA98-0DCB5EF7198A\",\"305ED6BA-094D-FACF-2012-73EE6E4650BE\",\"85A9A62B-5C57-318C-C391-D84BC8471369\",\"31203924-D33A-DBD4-3BEF-53BE4A7DD05A\"],\"qqwb\":{\"bind\":true,\"bind_time\":{\"sec\":1372815899,\"usec\":58000},\"access_token\":\"3907ef9e171fa611d73d589b53ddb987\",\"refresh_token\":\"f4a385701d905c2e0c16c1531c5c1699\",\"refresh_time\":{\"sec\":1372815899,\"usec\":58000},\"expires_in\":1380851097,\"openid\":\"59D54BA75BB7F09B965616F1FC0B6C65\",\"user_id\":\"59D54BA75BB7F09B965616F1FC0B6C65\",\"user_name\":\"lptmoon\",\"vip_desc\":\"\\u674e\\u5fd7\\u5a01\\uff0c\\u964c\\u964c\\u79d1\\u6280\\u9996\\u5e2d\\u6280\\u672f\\u5b98\\u3002\"},\"regtime\":{\"sec\":1344242533,\"usec\":127000},\"school\":\"\",\"sex\":\"M\",\"sign\":\"\\u2660\\u2665\\u2663\\u2666\",\"signex\":{\"time\":{\"sec\":1355671838,\"usec\":469000}},\"sina_weibo\":{\"bind\":true,\"bind_time\":{\"sec\":1377741643,\"usec\":521000},\"oauth_token\":\"2.008DoAoBl6GKOCfc3ef857a9fyzSsD\",\"user_id\":\"1655142045\",\"version\":2,\"expiretime\":\"7837157.999976\",\"remindtime\":\"7837157.999975\",\"vip_desc\":\"\\u964c\\u964c\\u79d1\\u6280\\u8054\\u5408\\u521b\\u59cb\\u4eba\\uff0cCTO\\u674e\\u5fd7\\u5a01\"},\"version\":10,\"website\":\"http:\\/\\/blog.latermoon.com\",\"_cache_src_\":\"api\",\"_cache_time_\":\"2013-09-04 00:10:41\",\"vip\":{\"expire\":1380470399,\"level\":1,\"start\":1377792000}}"

func thread(conn redis.Conn, count int, ch chan int) {
	t1 := time.Now()
	for i := 0; i < count; i++ {
		conn.Do("RPUSH", "list1", profileJson)
	}
	ch <- 1
	t2 := time.Now()
	fmt.Println("Done in:", t2.Sub(t1))
}

func main() {
	//host := ":6379"
	host := ":1603"

	chanCount := 10
	countPerThread := 10000
	clients := make([]redis.Conn, chanCount)
	ch := make(chan int, chanCount)
	for i := 0; i < chanCount; i++ {
		clients[i], _ = redis.Dial("tcp", host)
	}
	fmt.Println("start...")
	t1 := time.Now()
	for i := 0; i < chanCount; i++ {
		go thread(clients[i], countPerThread, ch)
	}
	for i := 0; i < chanCount; i++ {
		<-ch
	}
	elapsed := time.Now().Sub(t1)
	qps := float64(chanCount*countPerThread) / elapsed.Seconds()
	fmt.Println("count:", chanCount*countPerThread, "elapsed:", elapsed, "qps:", qps)
}

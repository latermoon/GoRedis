package leveltool

// /*
// _geo:user_f15m:count = 100
// _geo:user_f15m:value:100422:hash = 9rj5jgn5842p
// _geo:user_f15m:value:300000:hash = wx4g4hcvzfxg
// ...
// _geo:user_f15m:value:100422:time = 1378000907596
// _geo:user_f15m:value:300000:time = 1378002003135
// ...
// _geo:user_f15m:hash:9rj5jgn5842p:value:0 = 100422
// _geo:user_f15m:hash:wx4g4hcvzfxg:value:1 = 300000
// ...
// _geo:user_f15m:time:1378000907596:value:0 = 100422
// _geo:user_f15m:time:1378000907596:value:1 = 100428
// _geo:user_f15m:time:1378002003135:value:0 = 300000
// ...
// */

// import (
// 	"../geo"
// 	"github.com/syndtr/goleveldb/leveldb"
// 	"github.com/syndtr/goleveldb/leveldb/opt"
// 	"sync"
// 	"time"
// )

// type LevelGeo struct {
// 	db     *leveldb.DB
// 	ro     *opt.ReadOptions
// 	wo     *opt.WriteOptions
// 	prefix string
// 	mu     sync.Mutex
// }

// func NewLevelGeo(db *leveldb.DB, prefix string) (g *LevelGeo) {
// 	g = &LevelGeo{}
// 	g.db = db
// 	g.prefix = prefix
// 	return
// }

// // _geo:[prefix]:value:[.]:hash:[.]
// func (g *LevelGeo) hashKey(value string) (k []byte) {
// 	k = []byte("_geo:" + g.prefix + ":value:" + value + ":hash")
// 	return
// }

// // _geo:[prefix]:value:[.]:time:[.]
// func (g *LevelGeo) timeKey(value string) (k []byte) {
// 	k = []byte("_geo:" + g.prefix + ":value:" + value + ":time")
// 	return
// }

// // _geo:[prefix]:hash:[.]:value:[.]:[index]
// // _geo:user_f15m:hash:9rj5jgn5842p:value:0 = 100422
// func (g *LevelGeo) findHashValueKey(hash string, value string) (k []byte) {

// 	return
// }

// func (g *LevelGeo) findTimeValueKey(t string, value string) (k []byte) {
// 	return
// }

// func (g *LevelGeo) Insert(lat, lng float64, value string, t string) (err error) {
// 	g.mu.Lock()
// 	defer g.mu.Unlock()
// 	// 找出要操作的key
// 	hash := geo.Encode(lat, lng)
// 	hashkey := g.hashKey(value)
// 	timekey := g.timeKey(value)
// 	hashvaluekey := g.findHashValueKey(hash, value)
// 	timevaluekey := g.findTimeValueKey(t, value)
// 	// 批量更新
// 	batch := new(leveldb.Batch)
// 	batch.Put(hashkey, []byte(hash))
// 	batch.Put(timekey, []byte(time.Unix()))
// 	batch.Put(hashvaluekey, []byte(value))
// 	batch.Put(timevaluekey, []byte(value))
// 	err = g.db.Write(batch, g.wo)
// 	return
// }

// func (g *LevelGeo) RemoveValue(value string) (err error) {
// 	g.mu.Lock()
// 	defer g.mu.Unlock()
// 	// 判断是否存在
// 	hashkey := g.hashKey(value)
// 	var v []byte
// 	v, err = g.db.Get(hashkey, g.ro)
// 	if err != nil {
// 		return
// 	}
// 	if v != nil {
// 		hash := string(v)
// 		hashkey := g.hashKey(value)
// 		timekey := g.timeKey(value)
// 		var t []byte
// 		t, err = g.db.Get(timekey, g.ro)
// 		if err != nil {
// 			return
// 		}
// 		hashvaluekey := g.findHashValueKey(hash, value)
// 		timevaluekey := g.findTimeValueKey(string(t), value)
// 		// 批量更新
// 		batch := new(leveldb.Batch)
// 		batch.Delete(hashkey)
// 		batch.Delete(timekey)
// 		batch.Delete(hashvaluekey)
// 		batch.Delete(timevaluekey)
// 		err = g.db.Write(batch, g.wo)
// 	}
// }

// func (g *LevelGeo) SearchNearby(lat, lng float64) {

// }

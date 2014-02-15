package rocksdb

// #cgo LDFLAGS: -lrocksdb
// #include "rocksdb/c.h"
import "C"

func (o *Options) SetMaxBackgroundCompactions(n int) {
	C.rocksdb_options_set_max_background_compactions(o.Opt, C.int(n))
}

func (env *Env) SetBackgroundThreads(n int) {
	C.rocksdb_env_set_background_threads(env.Env, C.int(n))
}

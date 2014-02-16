package rocksdb

// #cgo LDFLAGS: -lrocksdb
// #include "rocksdb/c.h"
import "C"

func (o *Options) SetMaxBackgroundCompactions(n int) {
	C.rocksdb_options_set_max_background_compactions(o.Opt, C.int(n))
}

// SetBackgroundThreads sets the size of the thread pool used for
// compactions and memtable flushes.
func (env *Env) SetBackgroundThreads(n int) {
	C.rocksdb_env_set_background_threads(env.Env, C.int(n))
}

// SetHighPriorityBackgroundThreads sets the size of the high priority
// thread pool that can be used to prevent compactions from stalling
// memtable flushes.
func (env *Env) SetHighPriorityBackgroundThreads(n int) {
	C.rocksdb_env_set_high_priority_background_threads(env.Env, C.int(n))
}

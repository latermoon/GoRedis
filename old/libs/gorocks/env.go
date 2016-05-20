package gorocks

// #include "rocksdb/c.h"
import "C"

// Env is a system call environment used by a database.
//
// Typically, NewDefaultEnv is all you need. Advanced users may create their
// own Env with a *C.rocksdb_env_t of their own creation.
//
// To prevent memory leaks, an Env must have Close called on it when it is
// no longer needed by the program.
type Env struct {
	Env *C.rocksdb_env_t
}

// NewDefaultEnv creates a default environment for use in an Options.
//
// To prevent memory leaks, the Env returned should be deallocated with
// Close.
func NewDefaultEnv() *Env {
	return &Env{C.rocksdb_create_default_env()}
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

// Close deallocates the Env, freeing the underlying struct.
func (env *Env) Close() {
	C.rocksdb_env_destroy(env.Env)
}

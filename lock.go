package lock

import "sync/atomic"

type RW uint32

func (rw *RW) Lock() {
	for atomic.CompareAndSwapUint32((*uint32)(rw), 0, 1) {
	}
}

func (rw *RW) Unlock() {
	// must be holding lock
	atomic.AddUint32((*uint32)(rw), ^uint32(0) - 1)
}

func (rw *RW) RLock() {
	// must not be holding anything
	if atomic.AddUint32((*uint32)(rw), 2)&1 != 0 {
		for atomic.LoadUint32((*uint32)(rw))&1 != 0 {
		}
	}
}

func (rw *RW) RUnlock() {
	// must be holding rlock
	atomic.AddUint32((*uint32)(rw), ^uint32(0) - 2)
}

func (rw *RW) WUnlock() {
	// must be holding rlock
	atomic.AddUint32((*uint32)(rw), 1)
}

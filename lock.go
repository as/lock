// Package lock provides a downgradeable read/write spinlock
// supporting many concurrent readers or one writer. The writer
// can downgrade their write lock to a read lock.
//
// Behaviors:
// - Priority: readers over writers (grouped reads) over sequential writes.
// - Readers have priority over writers, reserving a read of the
//   value currently protected by the lock or the value currently
//  being written by a writer.
// - Writer can become a reader, releasing the write half of the lock
//
// Implementation details:
// Reader:
// - Increment the lock by +2, check for even result.
// - If not even, spin until it is even.
//   We know that the writer will eventually decrement the lock to even.
//   We also know no new writer takes the lock if it is not in a 0 state
//   (which we ensure in the first step by adding +2.
// - Lock acquired.
// - Unlock: To release the lock, we add -2.
//
// Writer:
// - CAS on the values [0, 1], write lock held if the CAS occurs.
// - Unlock: add -1.
// - Downgrade: add +1, writer is now a reader.
// - Downgrade unlock: add -2 (same as reader).
package lock

import "sync/atomic"

// RW is a downgradeable read/write spinlock.
type RW uint64

// Lock locks rw. If the lock is already in use, the calling goroutine
// spins until the rw is available.
func (rw *RW) Lock() {
	for !atomic.CompareAndSwapUint64((*uint64)(rw), 0, 1) {
	}
}

// Unlock unlocks rw. It is undefined if rw is not locked on entry
// to Unlock.
func (rw *RW) Unlock() {
	atomic.AddUint64((*uint64)(rw), (^uint64(0))-1)
}

// Lock locks rw for reading. If there is a concurrent writer
// the calling goroutine spins until the rw is available for
// reading.
func (rw *RW) RLock() {
	if atomic.AddUint64((*uint64)(rw), 2)&1 != 0 {
		for atomic.LoadUint64((*uint64)(rw))&1 != 0 {
		}
	}
}

// Unlock unlocks rw for reading. The operation is undefined if
// the read lock isn't held.
func (rw *RW) RUnlock() {
	atomic.AddUint64((*uint64)(rw), (^uint64(0))-2)
}

// Downgrade transitions rw from a write-locked state to a read-locked
// state. The caller must hold the write-locked state.
//
// Proper usage:
//
//  rw.Lock()
// /* write operation*/
//
//  rw.Downgrade()
//  /* now a reader: read operation */
//
//  rw.RUnlock()
//  /* release */
//
func (rw *RW) Downgrade() {
	atomic.AddUint64((*uint64)(rw), 1)
}

// Package lock provides a downgradeable read/write spinlock
// supporting many concurrent readers or one writer. The writer
// can downgrade their write lock to a read lock.
//
// Implementation details:
//
// Reader:
// - Increment the lock by 2 and ensure the result is even.
// - Decrement by 2 on unlock.
//
// Writer:
// 	- CAS on the values [0, 1], write lock held if the CAS occurs
//  - Decrement by 1 on unlock
//  - Downgrade by incrementing by 1. Writer -> Reader.
//  - Downgraded writer unlocks as a Reader decrementing by 2.
package lock

import "sync/atomic"

// RW is a downgradeable read/write spinlock.
type RW uint32

// Lock locks rw. If the lock is already in use, the calling goroutine
// spins until the rw is available.
func (rw *RW) Lock() {
	for atomic.CompareAndSwapUint32((*uint32)(rw), 0, 1) {
	}
}

// Unlock unlocks rw. It is undefined if rw is not locked on entry
// to Unlock.
func (rw *RW) Unlock() {
	atomic.AddUint32((*uint32)(rw), ^uint32(0)-1)
}

// Lock locks rw for reading. If there is a concurrent writer
// the calling goroutine spins until the rw is available for
// reading.
func (rw *RW) RLock() {
	if atomic.AddUint32((*uint32)(rw), 2)&1 != 0 {
		for atomic.LoadUint32((*uint32)(rw))&1 != 0 {
		}
	}
}

// Unlock unlocks rw for reading. The operation is undefined if
// the read lock isn't held.
func (rw *RW) RUnlock() {
	atomic.AddUint32((*uint32)(rw), ^uint32(0)-2)
}

// Downgrade transitions rw from a write-locked state to a read-locked
// state. The caller must hold the write-locked state.
//
// Proper usage:
//
// 	rw.Lock() 
// /* write operation*/
//
// 	rw.Downgrade()
//  /* now a reader: read operation */
//
//  rw.RUnlock()
//  /* release */
//
func (rw *RW) Downgrade() {
	atomic.AddUint32((*uint32)(rw), 1)
}

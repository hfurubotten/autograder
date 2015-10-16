package database

import (
	"strings"
	"sync"
)

// keyLocker allows for fine-grained locking on a per bucket/key pair.
type keyLocker struct {
	mu      *sync.Mutex          // protect map from concurrent updates
	keyLock map[string]*fineLock // map of fineLocks for each bucket/key pair
}

type fineLock struct {
	sync.Mutex
	cond   *sync.Cond
	locked bool // condition variable
}

func newKeyLocker() *keyLocker {
	return &keyLocker{
		mu:      &sync.Mutex{},
		keyLock: make(map[string]*fineLock),
	}
}

func (k *keyLocker) get(bucket, key string) (*fineLock, bool) {
	k.mu.Lock()
	defer k.mu.Unlock()
	bucketkey := strings.Join([]string{bucket, key}, "/")
	fl, ok := k.keyLock[bucketkey]
	if !ok {
		fl = &fineLock{}
		fl.cond = sync.NewCond(fl)
		fl.locked = false
		k.keyLock[bucketkey] = fl
	}
	return fl, ok
}

// lock acquires the lock on the given bucket/key pair to prevent concurrent
// access.  If the lock is already in use, the calling goroutine
// blocks until the mutex is available.
func (k *keyLocker) lock(bucket, key string) {
	wkl, _ := k.get(bucket, key)
	wkl.Lock()
	for wkl.locked { // wait for bucketkey to become unlocked
		wkl.cond.Wait()
	}
	wkl.locked = true
	wkl.Unlock()
}

// unlock releases the lock on the given bucket/key pair.
// It is a run-time error if the bucket/key pair is not locked on entry to unlock.
func (k *keyLocker) unlock(bucket string, key string) {
	wkl, ok := k.get(bucket, key)
	if !ok {
		panic("database: unlock of unlocked mutex on: " + bucket + "/" + key)
	}
	wkl.Lock()
	wkl.locked = false
	wkl.cond.Signal() // signal that bucketkey is unlocked
	wkl.Unlock()
	// delete(k.keyLock, bucketkey)
}

package database

import "sync"

var writerslock sync.Mutex
var writerkeys = make(map[string]map[string]valueLocker)

type valueLocker struct {
	sync.Mutex
	islocked bool
}

// Lock will lock a specified key in a bucket for further use.
func Lock(bucket string, key string) {
	writerslock.Lock()
	defer writerslock.Unlock()

	if _, ok := writerkeys[bucket]; !ok {
		writerkeys[bucket] = make(map[string]valueLocker)
	}

	if _, ok := writerkeys[bucket][key]; !ok {
		writerkeys[bucket][key] = valueLocker{}
	}

	wkl := writerkeys[bucket][key]
	wkl.Lock()
	wkl.islocked = true
	writerkeys[bucket][string(key)] = wkl
}

// Unlock will unlock a specified key in a bucket and make it usable for other
// tasks running.
func Unlock(bucket string, key string) {
	writerslock.Lock()
	defer writerslock.Unlock()

	if _, ok := writerkeys[bucket]; !ok {
		return
	}

	if _, ok := writerkeys[bucket][key]; !ok {
		return
	}

	wkl := writerkeys[bucket][key]
	if wkl.islocked {
		wkl.Unlock()
		wkl.islocked = false
	}
	writerkeys[bucket][key] = wkl
}

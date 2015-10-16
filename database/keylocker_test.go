package database

import "testing"

func TestLockPanic(t *testing.T) {
	keyLocks := newKeyLocker()
	defer func() {
		if r := recover(); r != nil {
			t.Log(r)
		}
	}()
	// test that unlock without prior call to lock causes panic
	keyLocks.unlock(tmpbucket, "some key")
	t.FailNow()
}

func TestLockFuncs(t *testing.T) {
	keyLocks := newKeyLocker()
	key := "hey dude"
	keyLocks.lock(tmpbucket, key)
	if _, hasKey := keyLocks.keyLock[tmpbucket+"/"+key]; !hasKey {
		t.Errorf("Expected lock on key '%s' not found", tmpbucket+"/"+key)
	}
	keyLocks.unlock(tmpbucket, key)
	if _, hasKey := keyLocks.keyLock[tmpbucket+"/"+key]; !hasKey {
		t.Errorf("Unexpected lock on key '%s' after unlock", tmpbucket+"/"+key)
	}

	// check that we can take lock on same bucket/key pair again
	keyLocks.lock(tmpbucket, key)
	if _, hasKey := keyLocks.keyLock[tmpbucket+"/"+key]; !hasKey {
		t.Errorf("Expected lock on key '%s' not found", tmpbucket+"/"+key)
	}
	keyLocks.unlock(tmpbucket, key)
	if _, hasKey := keyLocks.keyLock[tmpbucket+"/"+key]; !hasKey {
		t.Errorf("Unexpected lock on key '%s' after unlock", tmpbucket+"/"+key)
	}
}

func TestConcurrentLocking(t *testing.T) {
	keyLocks := newKeyLocker()
	key := "hey dude"
	done := make(chan bool, 3)
	go func(ch chan bool) {
		keyLocks.lock(tmpbucket, key)
		t.Log("I got the lock 1")
		keyLocks.unlock(tmpbucket, key)
		t.Log("I unlocked it  1")
		ch <- true
	}(done)
	go func(ch chan bool) {
		keyLocks.lock(tmpbucket, key)
		t.Log("I got the lock 2")
		keyLocks.unlock(tmpbucket, key)
		t.Log("I unlocked it  2")
		ch <- true
	}(done)
	go func(ch chan bool) {
		keyLocks.lock(tmpbucket, key)
		t.Log("I got the lock 3")
		keyLocks.unlock(tmpbucket, key)
		t.Log("I unlocked it  3")
		ch <- true
	}(done)
	<-done
	<-done
	<-done
}

package cache

import (
	"testing"
	"time"

	miniredis "github.com/alicebob/miniredis/v2"
)

func TestDistributedLockExclusive(t *testing.T) {
	// Start in-memory redis
	s, err := miniredis.Run()
	if err != nil {
		t.Fatalf("failed to start miniredis: %v", err)
	}
	defer s.Close()

	// Init client against miniredis
	if err := InitRedisClient(s.Addr(), "", "", false, 0); err != nil {
		t.Fatalf("InitRedisClient failed: %v", err)
	}
	defer Close()

	key := "test_lock_key"

	// 1) Acquire lock in goroutine A and hold it for a while
	doneA := make(chan struct{})
	go func() {
		if err := Lock(key, 5*time.Second, 100*time.Millisecond); err != nil {
			t.Errorf("goroutine A failed to acquire lock: %v", err)
			close(doneA)
			return
		}
		defer func() {
			if err := Unlock(key); err != nil {
				t.Errorf("goroutine A failed to unlock: %v", err)
			}
			close(doneA)
		}()
		// Hold the lock a bit so B times out
		time.Sleep(200 * time.Millisecond)
	}()

	// Ensure A has time to acquire
	time.Sleep(10 * time.Millisecond)

	// 2) Goroutine B should time out trying to get the same lock
	start := time.Now()
	if err := Lock(key, 5*time.Second, 50*time.Millisecond); err == nil {
		t.Fatalf("goroutine B unexpectedly acquired the lock")
	} else if err != ErrLockTimeout {
		t.Fatalf("goroutine B expected ErrLockTimeout, got: %v", err)
	}
	if elapsed := time.Since(start); elapsed < 50*time.Millisecond {
		t.Fatalf("goroutine B returned too quickly, elapsed=%s", elapsed)
	}

	// Wait for A to release
	<-doneA

	// 3) After release, we should be able to acquire quickly
	if err := Lock(key, 5*time.Second, 200*time.Millisecond); err != nil {
		t.Fatalf("failed to acquire after release: %v", err)
	}
	if err := Unlock(key); err != nil {
		t.Fatalf("failed to unlock after reacquire: %v", err)
	}
}
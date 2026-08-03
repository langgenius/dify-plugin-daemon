package helper

import (
	"fmt"
	"sync"
	"testing"
	"time"

	"github.com/langgenius/dify-plugin-daemon/pkg/entities/plugin_entities"
)

func newTestCache() *memCache {
	return &memCache{items: make(map[string]*memCacheItem)}
}

func declarationNamed(name string) *plugin_entities.PluginDeclaration {
	declaration := &plugin_entities.PluginDeclaration{}
	declaration.Name = name
	return declaration
}

// withinTimeout fails the test instead of hanging the suite when work never returns.
func withinTimeout(t *testing.T, timeout time.Duration, work func()) {
	t.Helper()

	done := make(chan struct{})
	go func() {
		defer close(done)
		work()
	}()

	select {
	case <-done:
	case <-time.After(timeout):
		t.Fatalf("work did not finish within %s", timeout)
	}
}

// Regression for the eviction loop spinning forever under the write lock: external
// deletes used to leave the size counter above the real map length, and once that
// drift reached maxMemCacheSize the loop could no longer evict its way out.
func TestSetTerminatesWhenEntriesAreDeletedExternally(t *testing.T) {
	c := newTestCache()

	withinTimeout(t, 30*time.Second, func() {
		for i := 0; i < maxMemCacheSize+64; i++ {
			key := fmt.Sprintf("key-%d", i)
			c.set(key, declarationNamed(fmt.Sprintf("declaration-%d", i)))

			c.Lock()
			delete(c.items, key)
			c.Unlock()
		}
	})
}

func TestSetTerminatesWhenOverwritingTheSameKey(t *testing.T) {
	c := newTestCache()

	withinTimeout(t, 30*time.Second, func() {
		for i := 0; i < maxMemCacheSize*4; i++ {
			c.set("stable-key", declarationNamed(fmt.Sprintf("declaration-%d", i)))
		}
	})

	c.RLock()
	size := len(c.items)
	c.RUnlock()

	if size != 1 {
		t.Fatalf("overwriting one key should leave one entry, got %d", size)
	}

	declaration := c.get("stable-key")
	if declaration == nil {
		t.Fatal("expected the overwritten key to still be cached")
	}
	if want := fmt.Sprintf("declaration-%d", maxMemCacheSize*4-1); declaration.Name != want {
		t.Fatalf("expected the last write to win, want %q got %q", want, declaration.Name)
	}
}

func TestSetEvictsToStayWithinMaxSize(t *testing.T) {
	c := newTestCache()

	withinTimeout(t, 60*time.Second, func() {
		for i := 0; i < maxMemCacheSize+128; i++ {
			c.set(fmt.Sprintf("key-%d", i), declarationNamed(fmt.Sprintf("declaration-%d", i)))
		}
	})

	c.RLock()
	size := len(c.items)
	c.RUnlock()

	if size > maxMemCacheSize {
		t.Fatalf("cache grew past maxMemCacheSize: got %d want <= %d", size, maxMemCacheSize)
	}
}

func TestGetEvictsExpiredEntries(t *testing.T) {
	c := newTestCache()
	c.set("expired-key", declarationNamed("expired-declaration"))

	c.Lock()
	c.items["expired-key"].lastAccess = time.Now().Add(-2 * maxTTL)
	c.Unlock()

	if declaration := c.get("expired-key"); declaration != nil {
		t.Fatalf("expected expired entry to be a miss, got %q", declaration.Name)
	}

	c.RLock()
	_, exists := c.items["expired-key"]
	c.RUnlock()

	if exists {
		t.Fatal("expected expired entry to be removed from the cache")
	}
}

// Exercises the concurrent overwrite pattern from CombinedGetPluginDeclaration, where
// several goroutines miss on the same key and all call set. Meaningful under -race.
func TestConcurrentGetSetAndDelete(t *testing.T) {
	c := newTestCache()

	withinTimeout(t, 60*time.Second, func() {
		var wg sync.WaitGroup
		for worker := 0; worker < 16; worker++ {
			wg.Add(1)
			go func(worker int) {
				defer wg.Done()
				for i := 0; i < 512; i++ {
					key := fmt.Sprintf("key-%d", i%32)
					if c.get(key) == nil {
						c.set(key, declarationNamed(fmt.Sprintf("declaration-%d-%d", worker, i)))
					}
					if i%64 == 0 {
						c.Lock()
						delete(c.items, key)
						c.Unlock()
					}
				}
			}(worker)
		}
		wg.Wait()
	})

	c.RLock()
	size := len(c.items)
	c.RUnlock()

	if size > maxMemCacheSize {
		t.Fatalf("cache grew past maxMemCacheSize: got %d want <= %d", size, maxMemCacheSize)
	}
}

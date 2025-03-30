package mapping

import (
	"sync"
	"testing"
)

// TestLoadStore validates basic read/write operations
func TestLoadStore(t *testing.T) {
	t.Parallel()
	m := Map[string, int]{}

	// Test initial state
	if val, ok := m.Load("missing"); ok {
		t.Fatalf("Unexpected value for missing key: %v", val)
	}

	// Test basic store
	m.Store("answer", 42)
	if val, ok := m.Load("answer"); !ok || val != 42 {
		t.Errorf("Load after Store failed, got (%v, %v)", val, ok)
	}

	// Test overwrite
	prevLen := m.Len()
	m.Store("answer", 100)
	if m.Len() != prevLen {
		t.Error("Overwriting existing key should not change length")
	}
}

// TestDelete validates deletion behavior
func TestDelete(t *testing.T) {
	t.Parallel()
	m := Map[string, string]{}

	// Delete non-existent key
	m.Delete("ghost")
	if m.Len() != 0 {
		t.Error("Deleting non-existent key should not affect length")
	}

	// Delete existing key
	m.Store("name", "gopher")
	m.Delete("name")
	if _, ok := m.Load("name"); ok || m.Len() != 0 {
		t.Error("Delete failed to remove item")
	}
}

// TestConcurrentAccess verifies thread safety
func TestConcurrentAccess(t *testing.T) {
	t.Parallel()
	m := Map[int, float64]{}
	const workers = 100

	var wg sync.WaitGroup
	wg.Add(workers)
	
	for i := 0; i < workers; i++ {
		go func(i int) {
			defer wg.Done()
			m.Store(i, float64(i)*1.5)
			m.Load(i)
			m.Delete(i)
		}(i)
	}
	wg.Wait()

	if m.Len() != 0 {
		t.Errorf("Expected empty map after concurrent ops, got len %d", m.Len())
	}
}

// TestLoadOrStore verifies conditional storage
func TestLoadOrStore(t *testing.T) {
	t.Parallel()
	m := Map[string, interface{}]{}

	// First store
	val, loaded := m.LoadOrStore("data", []byte{1,2,3})
	if loaded || val.([]byte)[0] != 1 {
		t.Error("Initial LoadOrStore failed")
	}

	// Existing key
	val, loaded = m.LoadOrStore("data", "new value")
	if !loaded || len(val.([]byte)) != 3 {
		t.Error("Existing key LoadOrStore failed")
	}
}



// TestEdgeCases covers special scenarios
func TestEdgeCases(t *testing.T) {
	t.Parallel()
	m := Map[bool, bool]{}

	// Zero value storage
	m.Store(true, false)
	if val, _ := m.Load(true); val != false {
		t.Error("Zero value storage failed")
	}

	// Clear operation
	m.Clear()
	if m.Len() != 0 {
		t.Error("Clear failed to reset map")
	}
}
package mapping

import (
	"sync"
	"sync/atomic"
)

type Map[K comparable, V any] struct {
	len   int32
	store sync.Map
	mu    sync.RWMutex
}

func (m *Map[K, V]) Load(key K) (value V, ok bool) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	v, ok := m.store.Load(key)
	if !ok {
		return
	}

	value, ok = v.(V)
	return
}

func (m *Map[K, V]) Store(key K, value V) {
	m.mu.Lock()
	defer m.mu.Unlock()
	// If the key already exists, we don't want to increment the length
	_, loaded := m.store.Load(key)
	if !loaded {
		atomic.AddInt32(&m.len, 1)
	}
	m.store.Store(key, value)
}

func (m *Map[K, V]) Delete(key K) {
	m.mu.Lock()
	defer m.mu.Unlock()
	_, loaded := m.store.Load(key)
	// If the key exists, we want to decrement the length
	// If the key does not exist, we don't want to decrement the length
	if loaded {
		atomic.AddInt32(&m.len, -1)
	}
	m.store.Delete(key)
}

func (m *Map[K, V]) Range(f func(key K, value V) bool) {
	m.store.Range(func(key, value interface{}) bool {
		return f(key.(K), value.(V))
	})
}

func (m *Map[K, V]) LoadOrStore(key K, value V) (actual V, loaded bool) {
	m.mu.Lock()
	defer m.mu.Unlock()
	
	v, loaded := m.store.LoadOrStore(key, value)
	actual = v.(V)
	if !loaded {
		atomic.AddInt32(&m.len, 1)
	}
	return
}

func (m *Map[K, V]) LoadAndDelete(key K) (value V, loaded bool) {
	m.mu.Lock()
	defer m.mu.Unlock()

	v, loaded := m.store.LoadAndDelete(key)
	value = v.(V)
	if loaded {
		atomic.AddInt32(&m.len, -1)
	}
	return
}

func (m *Map[K, V]) Swap(key K, value V) (actual V, swapped bool) {
	m.mu.Lock()
	defer m.mu.Unlock()

	v, swapped := m.store.Swap(key, value)
	actual = v.(V)
	return
}

func (m *Map[K, V]) Clear() {
	m.mu.Lock()
	defer m.mu.Unlock()
	// Clear the map
	m.store.Range(func(key, value interface{}) bool {
		m.store.Delete(key)
		return true
	})
	atomic.StoreInt32(&m.len, 0)
}

func (m *Map[K, V]) Len() int {
	return int(atomic.LoadInt32(&m.len))
}

func (m *Map[K, V]) Exists(key K) bool {
	_, ok := m.Load(key)
	return ok
}

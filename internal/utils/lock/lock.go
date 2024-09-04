package lock

import (
	"sync"
	"sync/atomic"
)

type mutex struct {
	*sync.Mutex
	count int32
}

type HighGranularityLock struct {
	m map[string]*mutex
	l sync.Mutex
}

func NewHighGranularityLock() *HighGranularityLock {
	return &HighGranularityLock{
		m: make(map[string]*mutex),
	}
}

func (l *HighGranularityLock) Lock(key string) {
	l.l.Lock()
	var m *mutex
	var ok bool
	if m, ok = l.m[key]; !ok {
		m = &mutex{Mutex: &sync.Mutex{}, count: 1}
		l.m[key] = m
	} else {
		atomic.AddInt32(&m.count, 1)
	}
	l.l.Unlock()

	m.Lock()
}

func (l *HighGranularityLock) Unlock(key string) {
	l.l.Lock()
	m, ok := l.m[key]
	if !ok {
		return
	}
	atomic.AddInt32(&m.count, -1)
	if atomic.LoadInt32(&m.count) == 0 {
		delete(l.m, key)
	}
	l.l.Unlock()
	m.Unlock()
}

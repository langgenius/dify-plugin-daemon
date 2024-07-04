package entities

import "sync"

type IOListener[T any] struct {
	l        *sync.RWMutex
	onClose  func()
	listener []func(T)
}

type BytesIOListener = IOListener[[]byte]

func NewIOListener[T any]() *IOListener[T] {
	return &IOListener[T]{
		l: &sync.RWMutex{},
	}
}

func (r *IOListener[T]) AddListener(f func(T)) {
	r.l.Lock()
	defer r.l.Unlock()
	r.listener = append(r.listener, f)
}

func (r *IOListener[T]) OnClose(f func()) {
	r.onClose = f
}

func (r *IOListener[T]) Close() {
	r.onClose()
}

func (r *IOListener[T]) Emit(data T) {
	r.l.RLock()
	defer r.l.RUnlock()
	for _, listener := range r.listener {
		listener(data)
	}
}

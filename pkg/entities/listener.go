package entities

import "sync"

type Broadcast[T any] struct {
	l        *sync.RWMutex
	onClose  []func()
	listener []func(T)
	closed   bool
}

type BytesIOListener = Broadcast[[]byte]

func NewCallbackHandler[T any]() *Broadcast[T] {
	return &Broadcast[T]{
		l: &sync.RWMutex{},
	}
}

func (r *Broadcast[T]) Listen(f func(T)) {
	r.l.Lock()
	defer r.l.Unlock()
	r.listener = append(r.listener, f)
}

func (r *Broadcast[T]) OnClose(f func()) {
	r.l.Lock()
	if r.closed {
		r.l.Unlock()
		f()
		return
	}
	r.onClose = append(r.onClose, f)
	r.l.Unlock()
}

func (r *Broadcast[T]) Close() {
	r.l.Lock()
	if r.closed {
		r.l.Unlock()
		return
	}
	r.closed = true
	onClose := append([]func(){}, r.onClose...)
	r.onClose = nil
	r.l.Unlock()

	for _, f := range onClose {
		f()
	}
}

func (r *Broadcast[T]) Send(data T) {
	r.l.RLock()
	listeners := append([]func(T){}, r.listener...)
	r.l.RUnlock()

	for _, listener := range listeners {
		listener(data)
	}
}

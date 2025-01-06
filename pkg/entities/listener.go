package entities

import "sync"

type Broadcast[T any] struct {
	l        *sync.RWMutex
	onClose  func()
	listener []func(T)
}

type BytesIOListener = Broadcast[[]byte]

func NewBroadcast[T any]() *Broadcast[T] {
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
	r.onClose = f
}

func (r *Broadcast[T]) Close() {
	if r.onClose != nil {
		r.onClose()
	}
}

func (r *Broadcast[T]) Send(data T) {
	r.l.RLock()
	defer r.l.RUnlock()
	for _, listener := range r.listener {
		listener(data)
	}
}

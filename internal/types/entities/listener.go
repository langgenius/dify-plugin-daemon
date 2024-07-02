package entities

import "sync/atomic"

type IOListener[T any] struct {
	msg        chan T
	closed     *int32
	close_hook []func()
}

type BytesIOListener = IOListener[[]byte]

func NewIOListener[T any]() *IOListener[T] {
	return &IOListener[T]{
		msg:        make(chan T),
		closed:     new(int32),
		close_hook: []func(){},
	}
}

func (r *IOListener[T]) Listen() <-chan T {
	return r.msg
}

func (r *IOListener[T]) Close() {
	if !atomic.CompareAndSwapInt32(r.closed, 0, 1) {
		return
	}
	atomic.StoreInt32(r.closed, 1)
	for _, hook := range r.close_hook {
		hook()
	}
	close(r.msg)
}

func (r *IOListener[T]) Write(data T) {
	if atomic.LoadInt32(r.closed) == 1 {
		return
	}
	r.msg <- data
}

func (r *IOListener[T]) OnClose(hook func()) {
	r.close_hook = append(r.close_hook, hook)
}

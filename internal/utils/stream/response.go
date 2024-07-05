package stream

import (
	"errors"
	"sync"

	"github.com/gammazero/deque"
)

type StreamResponse[T any] struct {
	q         deque.Deque[T]
	l         *sync.Mutex
	sig       chan bool
	closed    bool
	max       int
	listening bool
	onClose   func()
}

func NewStreamResponse[T any](max int) *StreamResponse[T] {
	return &StreamResponse[T]{
		l:   &sync.Mutex{},
		sig: make(chan bool),
		max: max,
	}
}

func (r *StreamResponse[T]) OnClose(f func()) {
	r.onClose = f
}

func (r *StreamResponse[T]) Next() bool {
	r.l.Lock()
	if r.closed {
		r.l.Unlock()
		return false
	}

	if r.q.Len() > 0 {
		r.l.Unlock()
		return true
	}

	r.listening = true
	defer func() {
		r.listening = false
	}()

	r.l.Unlock()
	return <-r.sig
}

func (r *StreamResponse[T]) Read() (T, error) {
	r.l.Lock()
	defer r.l.Unlock()

	if r.q.Len() > 0 {
		data := r.q.PopFront()
		return data, nil
	} else {
		var data T
		return data, errors.New("no data available, please call Next() to wait for data")
	}
}

func (r *StreamResponse[T]) Write(data T) error {
	r.l.Lock()
	if r.closed {
		r.l.Unlock()
		return nil
	}

	if r.q.Len() >= r.max {
		r.l.Unlock()
		return errors.New("queue is full")
	}

	r.q.PushBack(data)
	if r.q.Len() == 1 {
		if r.listening {
			r.sig <- true
		}
	}
	r.l.Unlock()
	return nil
}

func (r *StreamResponse[T]) Close() {
	r.l.Lock()
	if r.closed {
		r.l.Unlock()
		return
	}
	r.closed = true
	r.l.Unlock()

	select {
	case r.sig <- false:
	default:
	}
	close(r.sig)
	if r.onClose != nil {
		r.onClose()
	}
}

func (r *StreamResponse[T]) IsClosed() bool {
	r.l.Lock()
	defer r.l.Unlock()

	return r.closed
}

func (r *StreamResponse[T]) Size() int {
	r.l.Lock()
	defer r.l.Unlock()

	return r.q.Len()
}

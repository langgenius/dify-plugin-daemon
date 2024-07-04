package entities

import (
	"errors"
	"sync"

	"github.com/gammazero/deque"
)

type InvocationSession struct {
	ID             string
	PluginIdentity string
}

type InvocationResponse[T any] struct {
	q         deque.Deque[T]
	l         *sync.Mutex
	sig       chan bool
	closed    bool
	max       int
	listening bool
	onClose   func()
}

func NewInvocationResponse[T any](max int) *InvocationResponse[T] {
	return &InvocationResponse[T]{
		l:   &sync.Mutex{},
		sig: make(chan bool),
		max: max,
	}
}

func (r *InvocationResponse[T]) OnClose(f func()) {
	r.onClose = f
}

func (r *InvocationResponse[T]) Next() bool {
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

func (r *InvocationResponse[T]) Read() (T, error) {
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

func (r *InvocationResponse[T]) Write(data T) error {
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

func (r *InvocationResponse[T]) Close() {
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

func (r *InvocationResponse[T]) IsClosed() bool {
	r.l.Lock()
	defer r.l.Unlock()

	return r.closed
}

func (r *InvocationResponse[T]) Size() int {
	r.l.Lock()
	defer r.l.Unlock()

	return r.q.Len()
}

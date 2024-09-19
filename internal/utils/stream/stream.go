package stream

import (
	"errors"
	"sync"
	"sync/atomic"

	"github.com/gammazero/deque"
)

var ErrEmpty = errors.New("no data available")

type Stream[T any] struct {
	q         deque.Deque[T]
	l         *sync.Mutex
	sig       chan bool
	closed    int32
	max       int
	listening bool

	onClose     []func()
	beforeClose []func()
	filter      []func(T) error

	err error
}

func NewStream[T any](max int) *Stream[T] {
	return &Stream[T]{
		l:   &sync.Mutex{},
		sig: make(chan bool),
		max: max,
	}
}

// Filter filters the stream with a function
// if the function returns an error, the stream will be closed
func (r *Stream[T]) Filter(f func(T) error) {
	r.filter = append(r.filter, f)
}

// OnClose adds a function to be called when the stream is closed
func (r *Stream[T]) OnClose(f func()) {
	r.onClose = append(r.onClose, f)
}

// BeforeClose adds a function to be called before the stream is closed
func (r *Stream[T]) BeforeClose(f func()) {
	r.beforeClose = append(r.beforeClose, f)
}

// Next returns true if there are more data to be read
// and waits for the next data to be available
// returns false if the stream is closed
// NOTE: even if the stream is closed, it will return true if there is data available
func (r *Stream[T]) Next() bool {
	r.l.Lock()
	if r.closed == 1 && r.q.Len() == 0 && r.err == nil {
		r.l.Unlock()
		return false
	}

	if r.q.Len() > 0 || r.err != nil {
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

// Read reads buffered data from the stream and
// it returns error only if the buffer is empty or an error is written to the stream
func (r *Stream[T]) Read() (T, error) {
	r.l.Lock()
	defer r.l.Unlock()

	if r.q.Len() > 0 {
		data := r.q.PopFront()
		for _, f := range r.filter {
			err := f(data)
			if err != nil {
				// close the stream
				r.Close()
				return data, err
			}
		}
		return data, nil
	} else {
		var data T
		if r.err != nil {
			err := r.err
			r.err = nil
			return data, err
		}

		return data, ErrEmpty
	}
}

// Async wraps the stream with a new stream, and allows customized operations
func (r *Stream[T]) Async(fn func(T)) error {
	for r.Next() {
		data, err := r.Read()
		if err != nil {
			return err
		}
		fn(data)
	}

	return nil
}

// Write writes data to the stream,
// returns error if the buffer is full
func (r *Stream[T]) Write(data T) error {
	if atomic.LoadInt32(&r.closed) == 1 {
		return nil
	}

	r.l.Lock()

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

// Close closes the stream
func (r *Stream[T]) Close() {
	if !atomic.CompareAndSwapInt32(&r.closed, 0, 1) {
		return
	}

	for _, f := range r.beforeClose {
		f()
	}

	select {
	case r.sig <- false:
	default:
	}
	close(r.sig)

	for _, f := range r.onClose {
		f()
	}
}

func (r *Stream[T]) IsClosed() bool {
	return atomic.LoadInt32(&r.closed) == 1
}

func (r *Stream[T]) Size() int {
	r.l.Lock()
	defer r.l.Unlock()

	return r.q.Len()
}

// WriteError writes an error to the stream
func (r *Stream[T]) WriteError(err error) {
	if atomic.LoadInt32(&r.closed) == 1 {
		return
	}

	r.l.Lock()
	defer r.l.Unlock()

	r.err = err

	if r.q.Len() == 0 {
		if r.listening {
			r.sig <- true
		}
	}
}

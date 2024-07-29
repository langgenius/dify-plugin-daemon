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
	err       error
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

// Next returns true if there are more data to be read
// and waits for the next data to be available
// returns false if the stream is closed
// NOTE: even if the stream is closed, it will return true if there is data available
func (r *StreamResponse[T]) Next() bool {
	r.l.Lock()
	if r.closed && r.q.Len() == 0 && r.err == nil {
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
func (r *StreamResponse[T]) Read() (T, error) {
	r.l.Lock()
	defer r.l.Unlock()

	if r.q.Len() > 0 {
		data := r.q.PopFront()
		return data, nil
	} else {
		var data T
		if r.err != nil {
			err := r.err
			r.err = nil
			return data, err
		}

		return data, errors.New("no data available")
	}
}

// Wrap wraps the stream with a new stream, and allows customized operations
func (r *StreamResponse[T]) Wrap(fn func(T)) error {
	r.l.Lock()
	if r.closed {
		r.l.Unlock()
		return errors.New("stream is closed")
	}
	r.l.Unlock()

	for r.Next() {
		data, err := r.Read()
		if err != nil {
			return err
		}
		fn(data)
	}

	return nil
}

// Write writes data to the stream
// returns error if the buffer is full
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

// Close closes the stream
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

// WriteError writes an error to the stream
func (r *StreamResponse[T]) WriteError(err error) {
	r.l.Lock()
	defer r.l.Unlock()

	r.err = err

	if r.q.Len() == 0 {
		if r.listening {
			r.sig <- true
		}
	}
}

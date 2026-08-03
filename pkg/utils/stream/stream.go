package stream

import (
	"errors"
	"sync"
	"sync/atomic"

	"github.com/gammazero/deque"
)

var ErrEmpty = errors.New("no data available")

const (
	streamStateOpen int32 = iota
	streamStateClosing
	streamStateClosed
)

type Stream[T any] struct {
	q      deque.Deque[T]
	l      *sync.Mutex
	sig    chan struct{}
	done   chan struct{}
	closed int32
	max    int

	onClose     []func()
	beforeClose []func()
	filter      []func(T) error

	err error

	// Condition variable for blocking writes when queue is full
	writeCond *sync.Cond
}

func NewStream[T any](max int) *Stream[T] {
	mutex := &sync.Mutex{}
	return &Stream[T]{
		l:         mutex,
		sig:       make(chan struct{}, 1),
		done:      make(chan struct{}),
		max:       max,
		writeCond: sync.NewCond(mutex),
	}
}

// Filter filters the stream with a function
// if the function returns an error, the stream will be closed
func (r *Stream[T]) Filter(f func(T) error) {
	r.l.Lock()
	defer r.l.Unlock()
	if atomic.LoadInt32(&r.closed) != streamStateOpen {
		return
	}
	r.filter = append(r.filter, f)
}

// OnClose adds a function to be called when the stream is closed
func (r *Stream[T]) OnClose(f func()) {
	r.l.Lock()
	if atomic.LoadInt32(&r.closed) == streamStateClosed {
		r.l.Unlock()
		f()
		return
	}
	r.onClose = append(r.onClose, f)
	r.l.Unlock()
}

// BeforeClose adds a function to be called before the stream is closed
func (r *Stream[T]) BeforeClose(f func()) {
	r.l.Lock()
	defer r.l.Unlock()
	if atomic.LoadInt32(&r.closed) != streamStateOpen {
		return
	}
	r.beforeClose = append(r.beforeClose, f)
}

// Next returns true if there are more data to be read
// and waits for the next data to be available
// returns false if the stream is closed
// NOTE: even if the stream is closed, it will return true if there is data available
func (r *Stream[T]) Next() bool {
	for {
		r.l.Lock()
		state := atomic.LoadInt32(&r.closed)
		hasData := r.q.Len() > 0 || (r.err != nil && state != streamStateClosing)
		closed := state == streamStateClosed
		r.l.Unlock()
		if hasData {
			return true
		}
		if closed {
			return false
		}
		select {
		case <-r.sig:
		case <-r.done:
		}
	}
}

// Read reads buffered data from the stream and
// it returns error only if the buffer is empty or an error is written to the stream
func (r *Stream[T]) Read() (T, error) {
	r.l.Lock()

	if r.q.Len() > 0 {
		data := r.q.PopFront()
		// Signal any waiting writers that there's now space in the queue
		r.writeCond.Signal()
		filters := append([]func(T) error(nil), r.filter...)
		r.l.Unlock()

		for _, f := range filters {
			err := f(data)
			if err != nil {
				r.Close()
				return data, err
			}
		}
		return data, nil
	}

	var data T
	if r.err != nil && atomic.LoadInt32(&r.closed) != streamStateClosing {
		err := r.err
		r.err = nil
		r.l.Unlock()
		return data, err
	}
	r.l.Unlock()
	return data, ErrEmpty
}

// Process wraps the stream with a new stream, and allows customized operations
func (r *Stream[T]) Process(fn func(T)) error {
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
	if atomic.LoadInt32(&r.closed) != streamStateOpen {
		return nil
	}

	r.l.Lock()
	if atomic.LoadInt32(&r.closed) != streamStateOpen {
		r.l.Unlock()
		return nil
	}

	if r.q.Len() >= r.max {
		r.l.Unlock()
		return errors.New("queue is full")
	}

	r.q.PushBack(data)
	r.l.Unlock()
	r.signal()
	return nil
}

// WriteBlocking writes data to the stream,
// blocks if the buffer is full until space becomes available
func (r *Stream[T]) WriteBlocking(data T) {
	if atomic.LoadInt32(&r.closed) != streamStateOpen {
		return
	}

	r.l.Lock()

	// Wait until there's space in the queue or the stream is closed
	for r.q.Len() >= r.max && atomic.LoadInt32(&r.closed) == streamStateOpen {
		r.writeCond.Wait()
	}

	// Check if the stream was closed while waiting
	if atomic.LoadInt32(&r.closed) != streamStateOpen {
		r.l.Unlock()
		return
	}

	r.q.PushBack(data)
	r.l.Unlock()
	r.signal()
}

// Close closes the stream
func (r *Stream[T]) Close() {
	r.close(nil)
}

// CloseWithError closes the stream and makes err available to readers after close callbacks complete.
func (r *Stream[T]) CloseWithError(err error) {
	r.close(err)
}

func (r *Stream[T]) close(err error) {
	if !atomic.CompareAndSwapInt32(&r.closed, streamStateOpen, streamStateClosing) {
		return
	}

	r.l.Lock()
	if err != nil {
		r.err = err
	}
	beforeClose := append([]func(){}, r.beforeClose...)
	r.writeCond.Broadcast()
	r.l.Unlock()
	for _, f := range beforeClose {
		f()
	}

	for {
		r.l.Lock()
		if len(r.onClose) == 0 {
			atomic.StoreInt32(&r.closed, streamStateClosed)
			r.l.Unlock()
			close(r.done)
			return
		}
		onClose := append([]func(){}, r.onClose...)
		r.onClose = nil
		r.l.Unlock()

		for _, f := range onClose {
			f()
		}
	}
}

func (r *Stream[T]) IsClosed() bool {
	return atomic.LoadInt32(&r.closed) != streamStateOpen
}

func (r *Stream[T]) Size() int {
	r.l.Lock()
	defer r.l.Unlock()

	return r.q.Len()
}

// WriteError writes an error to the stream
func (r *Stream[T]) WriteError(err error) {
	if atomic.LoadInt32(&r.closed) != streamStateOpen {
		return
	}

	r.l.Lock()
	if atomic.LoadInt32(&r.closed) != streamStateOpen {
		r.l.Unlock()
		return
	}
	r.err = err
	r.l.Unlock()
	r.signal()
}

func (r *Stream[T]) signal() {
	select {
	case r.sig <- struct{}{}:
	default:
	}
}

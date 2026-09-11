package http_requests

import (
	"context"
	"fmt"
	"io"
	"sync"
	"time"
)

// StreamTimeouts separates transport progress from the total request budget.
// FirstResponse ends at the first decoded response frame, not the first LLM token.
type StreamTimeouts struct {
	FirstResponse time.Duration
	ReadIdle      time.Duration
	Total         time.Duration
}

type StreamTimeoutPhase string

const (
	StreamTimeoutFirstResponse StreamTimeoutPhase = "first_response"
	StreamTimeoutReadIdle      StreamTimeoutPhase = "read_idle"
	StreamTimeoutTotal         StreamTimeoutPhase = "total"
)

type StreamTimeoutError struct {
	Phase    StreamTimeoutPhase
	Duration time.Duration
}

func (e *StreamTimeoutError) Error() string {
	return fmt.Sprintf("stream invocation timed out (phase=%s, timeout_ms=%d)", e.Phase, e.Duration.Milliseconds())
}

func (e *StreamTimeoutError) Unwrap() error { return context.DeadlineExceeded }
func (e *StreamTimeoutError) Timeout() bool { return true }

type streamWatchdog struct {
	ctx    context.Context
	cancel context.CancelCauseFunc
	limits StreamTimeouts
	start  time.Time

	mu        sync.Mutex
	timer     *time.Timer
	responded bool
	lastRead  time.Time
	stopped   bool
	cause     error
}

func newStreamWatchdog(parent context.Context, limits StreamTimeouts) (*streamWatchdog, error) {
	if limits.FirstResponse <= 0 || limits.ReadIdle <= 0 || limits.Total <= 0 {
		return nil, fmt.Errorf("stream first response, read idle, and total timeouts must be positive")
	}
	ctx, cancel := context.WithCancelCause(parent)
	w := &streamWatchdog{ctx: ctx, cancel: cancel, limits: limits, start: time.Now()}
	w.mu.Lock()
	deadline, _ := w.deadlineLocked()
	w.timer = time.AfterFunc(time.Until(deadline), w.expire)
	w.mu.Unlock()
	return w, nil
}

func (w *streamWatchdog) deadlineLocked() (time.Time, *StreamTimeoutError) {
	deadline := w.start.Add(w.limits.Total)
	err := &StreamTimeoutError{Phase: StreamTimeoutTotal, Duration: w.limits.Total}
	phaseDeadline := w.start.Add(w.limits.FirstResponse)
	phase, timeout := StreamTimeoutFirstResponse, w.limits.FirstResponse
	if w.responded {
		phaseDeadline = w.lastRead.Add(w.limits.ReadIdle)
		phase, timeout = StreamTimeoutReadIdle, w.limits.ReadIdle
	}
	if phaseDeadline.Before(deadline) {
		return phaseDeadline, &StreamTimeoutError{Phase: phase, Duration: timeout}
	}
	return deadline, err
}

func (w *streamWatchdog) expire() {
	w.mu.Lock()
	if w.stopped || w.ctx.Err() != nil {
		w.mu.Unlock()
		return
	}
	deadline, cause := w.deadlineLocked()
	if remaining := time.Until(deadline); remaining > 0 {
		// A previous timer callback may already be running when progress resets it.
		w.timer.Reset(remaining)
		w.mu.Unlock()
		return
	}
	w.stopped, w.cause = true, cause
	w.mu.Unlock()
	w.cancel(cause)
}

func (w *streamWatchdog) progress(response bool) error {
	w.mu.Lock()
	if w.cause != nil {
		cause := w.cause
		w.mu.Unlock()
		return cause
	}
	if w.stopped || w.ctx.Err() != nil {
		cause := context.Cause(w.ctx)
		w.mu.Unlock()
		return cause
	}
	now := time.Now()
	deadline, cause := w.deadlineLocked()
	if !now.Before(deadline) {
		w.stopped, w.cause = true, cause
		w.timer.Stop()
		w.mu.Unlock()
		w.cancel(cause)
		return cause
	}
	if response {
		w.responded = true
	}
	if w.responded {
		w.lastRead = now
	}
	deadline, _ = w.deadlineLocked()
	w.timer.Reset(time.Until(deadline))
	w.mu.Unlock()
	return nil
}

// stop returns the real terminal cause before cancelling for resource cleanup.
func (w *streamWatchdog) stop() error {
	w.mu.Lock()
	w.stopped = true
	w.timer.Stop()
	cause := w.cause
	if cause == nil {
		cause = context.Cause(w.ctx)
	}
	w.mu.Unlock()
	w.cancel(nil)
	return cause
}

type streamActivityReader struct {
	io.Reader
	watchdog *streamWatchdog
}

func (r *streamActivityReader) Read(p []byte) (int, error) {
	n, err := r.Reader.Read(p)
	if n > 0 {
		if cause := r.watchdog.progress(false); cause != nil {
			return n, cause
		}
	}
	return n, err
}

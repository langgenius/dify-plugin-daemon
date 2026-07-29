package entities

import (
	"sync/atomic"
	"testing"
	"time"
)

func TestBroadcastCloseRunsAllCallbacksOnce(t *testing.T) {
	listener := NewCallbackHandler[int]()
	var calls atomic.Int32

	listener.OnClose(func() {
		calls.Add(1)
	})
	listener.OnClose(func() {
		calls.Add(1)
	})

	listener.Close()
	listener.Close()

	if got := calls.Load(); got != 2 {
		t.Fatalf("expected both callbacks to run once, got %d calls", got)
	}
}

func TestBroadcastOnCloseAddedAfterCloseRunsImmediately(t *testing.T) {
	listener := NewCallbackHandler[int]()
	listener.Close()

	var called atomic.Bool
	listener.OnClose(func() {
		called.Store(true)
	})

	if !called.Load() {
		t.Fatal("expected callback added after close to run immediately")
	}
}

func TestBroadcastListenerCanCloseFromCallback(t *testing.T) {
	listener := NewCallbackHandler[int]()
	done := make(chan struct{})
	listener.Listen(func(int) {
		listener.Close()
		close(done)
	})

	go listener.Send(1)

	select {
	case <-done:
	case <-time.After(time.Second):
		t.Fatal("listener callback deadlocked while closing broadcast")
	}
}

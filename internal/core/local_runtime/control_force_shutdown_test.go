package local_runtime

import (
	"io"
	"sync"
	"testing"
	"time"

	"github.com/langgenius/dify-plugin-daemon/internal/types/app"
)

// fake write closer captures Close calls
type fakeWC struct {
	onClose func()
	closed  bool
}

func (f *fakeWC) Write(p []byte) (int, error) { return len(p), nil }
func (f *fakeWC) Close() error {
	if !f.closed {
		f.closed = true
		if f.onClose != nil {
			f.onClose()
		}
	}
	return nil
}

// fake read closer returns EOF immediately and triggers on Close
type fakeRC struct {
	onClose func()
	closed  bool
}

func (f *fakeRC) Read(p []byte) (int, error) { return 0, io.EOF }
func (f *fakeRC) Close() error {
	if !f.closed {
		f.closed = true
		if f.onClose != nil {
			f.onClose()
		}
	}
	return nil
}

func TestForcefullyShutdownAllInstances(t *testing.T) {
	r := &LocalPluginRuntime{
		instances:      []*PluginInstance{},
		instanceLocker: &sync.RWMutex{},
		appConfig:      &app.Config{PluginMaxExecutionTimeout: 1},
	}

	var mu sync.Mutex
	closedInWriters := 0

	makeInst := func() *PluginInstance {
		wc := &fakeWC{onClose: func() {
			mu.Lock()
			closedInWriters++
			mu.Unlock()
		}} 
		// When stdout is closed, simulate runtime removing the first instance
		rc := &fakeRC{onClose: func() {
			r.instanceLocker.Lock()
			if len(r.instances) > 0 {
				// remove head to mimic OnInstanceShutdown handler
				r.instances = r.instances[1:]
			}
			r.instanceLocker.Unlock()
		}} 
		ec := &fakeRC{}
		return &PluginInstance{
			inWriter:  wc,
			outReader: rc,
			errReader: ec,
			l:         &sync.Mutex{},
			appConfig: &app.Config{PluginRuntimeBufferSize: 1024, PluginRuntimeMaxBufferSize: 1024},
		}
	}

	// seed two instances
	r.instances = append(r.instances, makeInst(), makeInst())

	start := time.Now()
	r.forcefullyShutdownAllInstances()
	elapsed := time.Since(start)

	if got := len(r.instances); got != 0 {
		t.Fatalf("expected all instances removed, got %d remaining", got)
	}
	if closedInWriters < 2 {
		t.Fatalf("expected Close to be called on all inWriters, got %d", closedInWriters)
	}
	// sanity: the loop sleeps 1s per iteration, so with 2 instances it should take at least ~2s
	if elapsed < 1500*time.Millisecond {
		t.Errorf("expected function to take at least ~1.5s due to sleeps, took %v", elapsed)
	}
}

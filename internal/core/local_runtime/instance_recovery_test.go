package local_runtime

import (
	"errors"
	"fmt"
	"io"
	"os"
	"sync"
	"syscall"
	"testing"
)

func TestIsInstanceDeadErr(t *testing.T) {
	tests := []struct {
		name string
		err  error
		want bool
	}{
		{name: "nil", err: nil, want: false},
		{name: "instance stopped", err: ErrInstanceStopped, want: true},
		{name: "wrapped closed file", err: fmt.Errorf("write stdin: %w", os.ErrClosed), want: true},
		{name: "closed pipe", err: io.ErrClosedPipe, want: true},
		{name: "short write", err: io.ErrShortWrite, want: true},
		{name: "broken pipe", err: syscall.EPIPE, want: true},
		{name: "closed file message", err: errors.New("write |1: file already closed"), want: true},
		{name: "other error", err: errors.New("temporary network timeout"), want: false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := IsInstanceDeadErr(tt.err); got != tt.want {
				t.Fatalf("IsInstanceDeadErr() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestPickLowestLoadInstanceSkipsStoppedInstances(t *testing.T) {
	stopped := &PluginInstance{instanceId: "stopped"}
	stopped.stopped.Store(true)
	live := &PluginInstance{instanceId: "live"}
	runtime := &LocalPluginRuntime{
		instances:      []*PluginInstance{stopped, live},
		instanceLocker: &sync.RWMutex{},
	}

	for i := 0; i < 4; i++ {
		instance, err := runtime.pickLowestLoadInstance()
		if err != nil {
			t.Fatalf("pickLowestLoadInstance() error = %v", err)
		}
		if instance != live {
			t.Fatalf("pickLowestLoadInstance() = %q, want live instance", instance.instanceId)
		}
	}
}

func TestPickLowestLoadInstanceReturnsErrWhenAllInstancesStopped(t *testing.T) {
	stopped := &PluginInstance{instanceId: "stopped"}
	stopped.stopped.Store(true)
	runtime := &LocalPluginRuntime{
		instances:      []*PluginInstance{stopped},
		instanceLocker: &sync.RWMutex{},
	}

	if _, err := runtime.pickLowestLoadInstance(); !errors.Is(err, ErrNoProperInstance) {
		t.Fatalf("pickLowestLoadInstance() error = %v, want %v", err, ErrNoProperInstance)
	}
}

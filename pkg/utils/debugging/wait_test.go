package debugging

import (
	"testing"
	"time"
)

func TestPossibleBlocking(t *testing.T) {
	triggered := false

	PossibleBlocking(func() any {
		time.Sleep(time.Second * 1)
		return nil
	}, time.Millisecond*500, func() {
		triggered = true
	})

	if !triggered {
		t.Fatal("possible blocking not triggered")
	}
}

func TestPossibleBlocking_Blocking(t *testing.T) {
	triggered := false

	PossibleBlocking(func() any {
		return nil
	}, time.Second*1, func() {
		triggered = true
	})

	if triggered {
		t.Fatal("possible blocking triggered")
	}
}

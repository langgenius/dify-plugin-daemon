package plugin_entities

import (
	"testing"
	"time"
)

func TestRuntimeStateHash(t *testing.T) {
	state := PluginRuntimeState{
		Restarts:  0,
		Status:    PLUGIN_RUNTIME_STATUS_PENDING,
		ActiveAt:  &[]time.Time{time.Now()}[0],
		StoppedAt: &[]time.Time{time.Now()}[0],
		Verified:  true,
	}

	hash, err := state.Hash()
	if err != nil {
		t.Errorf("hash failed: %v", err)
		return
	}

	if hash == 0 {
		t.Errorf("hash is 0")
		return
	}

	hash2, err := state.Hash()
	if err != nil {
		t.Errorf("hash failed: %v", err)
		return
	}

	if hash != hash2 {
		t.Errorf("hash is not the same: %d, %d", hash, hash2)
		return
	}

	state.Restarts++
	hash3, err := state.Hash()
	if err != nil {
		t.Errorf("hash failed: %v", err)
		return
	}

	if hash == hash3 {
		t.Errorf("hash is the same: %d, %d", hash, hash3)
		return
	}
}

package controlpanel

import (
	"testing"

	"github.com/langgenius/dify-plugin-daemon/internal/core/local_runtime"
	"github.com/langgenius/dify-plugin-daemon/internal/types/app"
	"github.com/langgenius/dify-plugin-daemon/pkg/entities/plugin_entities"
)

func TestLocalReadinessSnapshot(t *testing.T) {
	cfg := &app.Config{
		Platform:                       app.PLATFORM_LOCAL,
		PluginLocalLaunchingConcurrent: 1,
	}
	cp := NewControlPanel(cfg, nil, nil, nil)

	id, err := plugin_entities.NewPluginUniqueIdentifier("langgenius/test:1.0.0@1234567890abcdef1234567890abcdef")
	if err != nil {
		t.Fatalf("NewPluginUniqueIdentifier() error: %v", err)
	}

	cp.updateLocalReadinessSnapshot([]plugin_entities.PluginUniqueIdentifier{id})
	s1, ok := cp.LocalReadiness()
	if !ok {
		t.Fatalf("LocalReadiness() should have snapshot")
	}
	if s1.Ready {
		t.Fatalf("expected unready when runtime is missing")
	}
	if len(s1.Missing) != 1 {
		t.Fatalf("expected 1 missing plugin, got %d", len(s1.Missing))
	}

	cp.localPluginRuntimes.Store(id, &local_runtime.LocalPluginRuntime{})
	cp.updateLocalReadinessSnapshot([]plugin_entities.PluginUniqueIdentifier{id})
	s2, _ := cp.LocalReadiness()
	if !s2.Ready {
		t.Fatalf("expected ready after runtime exists")
	}
	if len(s2.Missing) != 0 || len(s2.Failed) != 0 {
		t.Fatalf("expected no missing/failed plugins")
	}
}

package tasks

import (
	"errors"
	"testing"
	"time"

	"github.com/langgenius/dify-plugin-daemon/internal/types/models"
	"github.com/langgenius/dify-plugin-daemon/internal/types/models/curd"
	"github.com/langgenius/dify-plugin-daemon/pkg/entities/plugin_entities"
)

var ErrTestRemoveLocalPlugin = errors.New("remove local plugin failed")
var ErrTestGracefulShutdown = errors.New("graceful shutdown failed")
var ErrTestForcefulShutdown = errors.New("forceful shutdown failed")

type mockPluginShutdownManager struct {
	removeLocalPluginFn             func(plugin_entities.PluginUniqueIdentifier) error
	shutdownLocalPluginGracefullyFn func(plugin_entities.PluginUniqueIdentifier) (<-chan error, error)
	shutdownLocalPluginForcefullyFn func(plugin_entities.PluginUniqueIdentifier) (<-chan error, error)
}

func (m *mockPluginShutdownManager) RemoveLocalPlugin(id plugin_entities.PluginUniqueIdentifier) error {
	if m.removeLocalPluginFn != nil {
		return m.removeLocalPluginFn(id)
	}
	return nil
}

func (m *mockPluginShutdownManager) ShutdownLocalPluginGracefully(id plugin_entities.PluginUniqueIdentifier) (<-chan error, error) {
	if m.shutdownLocalPluginGracefullyFn != nil {
		return m.shutdownLocalPluginGracefullyFn(id)
	}
	ch := make(chan error)
	close(ch)
	return ch, nil
}

func (m *mockPluginShutdownManager) ShutdownLocalPluginForcefully(id plugin_entities.PluginUniqueIdentifier) (<-chan error, error) {
	if m.shutdownLocalPluginForcefullyFn != nil {
		return m.shutdownLocalPluginForcefullyFn(id)
	}
	ch := make(chan error)
	close(ch)
	return ch, nil
}

func localPluginResponse() *curd.UpgradePluginResponse {
	return &curd.UpgradePluginResponse{
		IsOriginalPluginDeleted: true,
		DeletedPlugin: &models.Plugin{
			InstallType: plugin_entities.PLUGIN_RUNTIME_TYPE_LOCAL,
		},
	}
}

func mustParsePluginUniqueIdentifier(t *testing.T, s string) plugin_entities.PluginUniqueIdentifier {
	t.Helper()
	id, err := plugin_entities.NewPluginUniqueIdentifier(s)
	if err != nil {
		t.Fatalf("failed to parse plugin unique identifier %q: %v", s, err)
	}
	return id
}

// testPluginID is a valid plugin unique identifier for testing
const testPluginID = "langgenius/jina:0.0.4@a1b2c3d4e5f6a1b2c3d4e5f6a1b2c3d4"

func TestRemovePluginIfNeeded_IsOriginalPluginDeletedFalse(t *testing.T) {
	manager := &mockPluginShutdownManager{}
	response := &curd.UpgradePluginResponse{
		IsOriginalPluginDeleted: false,
	}
	pluginID := mustParsePluginUniqueIdentifier(t, testPluginID)

	err := RemovePluginIfNeeded(manager, pluginID, response)
	if err != nil {
		t.Errorf("expected nil error, got: %v", err)
	}
}

func TestRemovePluginIfNeeded_DeletedPluginNil(t *testing.T) {
	manager := &mockPluginShutdownManager{}
	response := &curd.UpgradePluginResponse{
		IsOriginalPluginDeleted: true,
		DeletedPlugin:           nil,
	}
	pluginID := mustParsePluginUniqueIdentifier(t, testPluginID)

	err := RemovePluginIfNeeded(manager, pluginID, response)
	if err != nil {
		t.Errorf("expected nil error, got: %v", err)
	}
}

func TestRemovePluginIfNeeded_NotLocalRuntime(t *testing.T) {
	manager := &mockPluginShutdownManager{}
	response := &curd.UpgradePluginResponse{
		IsOriginalPluginDeleted: true,
		DeletedPlugin: &models.Plugin{
			InstallType: plugin_entities.PLUGIN_RUNTIME_TYPE_SERVERLESS,
		},
	}
	pluginID := mustParsePluginUniqueIdentifier(t, testPluginID)

	err := RemovePluginIfNeeded(manager, pluginID, response)
	if err != nil {
		t.Errorf("expected nil error, got: %v", err)
	}
}

func TestRemovePluginIfNeeded_RemoveLocalPluginFails(t *testing.T) {
	manager := &mockPluginShutdownManager{
		removeLocalPluginFn: func(_ plugin_entities.PluginUniqueIdentifier) error {
			return ErrTestRemoveLocalPlugin
		},
	}
	pluginID := mustParsePluginUniqueIdentifier(t, testPluginID)

	err := RemovePluginIfNeeded(manager, pluginID, localPluginResponse())
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !errors.Is(err, ErrTestRemoveLocalPlugin) {
		t.Errorf("expected error to wrap ErrTestRemoveLocalPlugin, got: %v", err)
	}
}

func TestRemovePluginIfNeeded_GracefulShutdownSucceeds(t *testing.T) {
	manager := &mockPluginShutdownManager{}
	pluginID := mustParsePluginUniqueIdentifier(t, testPluginID)

	err := RemovePluginIfNeeded(manager, pluginID, localPluginResponse())
	if err != nil {
		t.Errorf("expected nil error, got: %v", err)
	}
}

func TestRemovePluginIfNeeded_GracefulShutdownChannelCloses(t *testing.T) {
	ch := make(chan error)
	go func() {
		time.Sleep(10 * time.Millisecond)
		close(ch)
	}()

	manager := &mockPluginShutdownManager{
		shutdownLocalPluginGracefullyFn: func(_ plugin_entities.PluginUniqueIdentifier) (<-chan error, error) {
			return ch, nil
		},
	}
	pluginID := mustParsePluginUniqueIdentifier(t, testPluginID)

	err := RemovePluginIfNeeded(manager, pluginID, localPluginResponse())
	if err != nil {
		t.Errorf("expected nil error, got: %v", err)
	}
}

func TestRemovePluginIfNeeded_GracefulShutdownReturnsNil(t *testing.T) {
	manager := &mockPluginShutdownManager{
		shutdownLocalPluginGracefullyFn: func(_ plugin_entities.PluginUniqueIdentifier) (<-chan error, error) {
			return nil, nil
		},
	}
	pluginID := mustParsePluginUniqueIdentifier(t, testPluginID)

	err := RemovePluginIfNeeded(manager, pluginID, localPluginResponse())
	if err != nil {
		t.Errorf("expected nil error, got: %v", err)
	}
}

func TestRemovePluginIfNeeded_GracefulFails_ForcefulSucceeds(t *testing.T) {
	manager := &mockPluginShutdownManager{
		shutdownLocalPluginGracefullyFn: func(_ plugin_entities.PluginUniqueIdentifier) (<-chan error, error) {
			return nil, ErrTestGracefulShutdown
		},
	}
	pluginID := mustParsePluginUniqueIdentifier(t, testPluginID)

	err := RemovePluginIfNeeded(manager, pluginID, localPluginResponse())
	if err != nil {
		t.Errorf("expected nil error, got: %v", err)
	}
}

func TestRemovePluginIfNeeded_GracefulFails_ForcefulAlsoFails(t *testing.T) {
	manager := &mockPluginShutdownManager{
		shutdownLocalPluginGracefullyFn: func(_ plugin_entities.PluginUniqueIdentifier) (<-chan error, error) {
			return nil, ErrTestGracefulShutdown
		},
		shutdownLocalPluginForcefullyFn: func(_ plugin_entities.PluginUniqueIdentifier) (<-chan error, error) {
			return nil, ErrTestForcefulShutdown
		},
	}
	pluginID := mustParsePluginUniqueIdentifier(t, testPluginID)

	err := RemovePluginIfNeeded(manager, pluginID, localPluginResponse())
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !errors.Is(err, ErrTestForcefulShutdown) {
		t.Errorf("expected error to wrap ErrTestForcefulShutdown, got: %v", err)
	}
}

func TestRemovePluginIfNeeded_GracefulTimesOut_ForcefulSucceeds(t *testing.T) {
	origTimeout := gracefulShutdownTimeout
	gracefulShutdownTimeout = 50 * time.Millisecond
	defer func() { gracefulShutdownTimeout = origTimeout }()

	gracefulCh := make(chan error) // never closes, will trigger timeout

	forceCh := make(chan error)
	go func() {
		time.Sleep(10 * time.Millisecond)
		close(forceCh)
	}()

	manager := &mockPluginShutdownManager{
		shutdownLocalPluginGracefullyFn: func(_ plugin_entities.PluginUniqueIdentifier) (<-chan error, error) {
			return gracefulCh, nil
		},
		shutdownLocalPluginForcefullyFn: func(_ plugin_entities.PluginUniqueIdentifier) (<-chan error, error) {
			return forceCh, nil
		},
	}
	pluginID := mustParsePluginUniqueIdentifier(t, testPluginID)

	err := RemovePluginIfNeeded(manager, pluginID, localPluginResponse())
	if err != nil {
		t.Errorf("expected nil error, got: %v", err)
	}
}

func TestRemovePluginIfNeeded_GracefulTimesOut_ForcefulTimesOut(t *testing.T) {
	origTimeout := gracefulShutdownTimeout
	gracefulShutdownTimeout = 50 * time.Millisecond
	defer func() { gracefulShutdownTimeout = origTimeout }()

	gracefulCh := make(chan error) // never closes
	forceCh := make(chan error)    // never closes

	manager := &mockPluginShutdownManager{
		shutdownLocalPluginGracefullyFn: func(_ plugin_entities.PluginUniqueIdentifier) (<-chan error, error) {
			return gracefulCh, nil
		},
		shutdownLocalPluginForcefullyFn: func(_ plugin_entities.PluginUniqueIdentifier) (<-chan error, error) {
			return forceCh, nil
		},
	}
	pluginID := mustParsePluginUniqueIdentifier(t, testPluginID)

	err := RemovePluginIfNeeded(manager, pluginID, localPluginResponse())
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if err.Error() != "forceful shutdown timed out" {
		t.Errorf("expected 'forceful shutdown timed out' error, got: %v", err)
	}
}

func TestRemovePluginIfNeeded_GracefulFails_ForcefulChannelCloses(t *testing.T) {
	forceCh := make(chan error)
	go func() {
		time.Sleep(10 * time.Millisecond)
		close(forceCh)
	}()

	manager := &mockPluginShutdownManager{
		shutdownLocalPluginGracefullyFn: func(_ plugin_entities.PluginUniqueIdentifier) (<-chan error, error) {
			return nil, ErrTestGracefulShutdown
		},
		shutdownLocalPluginForcefullyFn: func(_ plugin_entities.PluginUniqueIdentifier) (<-chan error, error) {
			return forceCh, nil
		},
	}
	pluginID := mustParsePluginUniqueIdentifier(t, testPluginID)

	err := RemovePluginIfNeeded(manager, pluginID, localPluginResponse())
	if err != nil {
		t.Errorf("expected nil error, got: %v", err)
	}
}

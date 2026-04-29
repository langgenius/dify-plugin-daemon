package controlpanel

import (
	"reflect"
	"testing"
	"unsafe"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"

	"github.com/langgenius/dify-plugin-daemon/internal/core/debugging_runtime"
	"github.com/langgenius/dify-plugin-daemon/internal/core/local_runtime"
	"github.com/langgenius/dify-plugin-daemon/internal/types/app"
	"github.com/langgenius/dify-plugin-daemon/pkg/entities/manifest_entities"
	"github.com/langgenius/dify-plugin-daemon/pkg/entities/plugin_entities"
)

type mockNotifier struct {
	connected    bool
	disconnected bool
}

func (m *mockNotifier) OnLocalRuntimeStarting(plugin_entities.PluginUniqueIdentifier)           {}
func (m *mockNotifier) OnLocalRuntimeReady(*local_runtime.LocalPluginRuntime)                   {}
func (m *mockNotifier) OnLocalRuntimeStartFailed(plugin_entities.PluginUniqueIdentifier, error) {}
func (m *mockNotifier) OnLocalRuntimeStop(*local_runtime.LocalPluginRuntime)                    {}
func (m *mockNotifier) OnLocalRuntimeStopped(*local_runtime.LocalPluginRuntime)                 {}
func (m *mockNotifier) OnLocalRuntimeScaleUp(*local_runtime.LocalPluginRuntime, int32)          {}
func (m *mockNotifier) OnLocalRuntimeScaleDown(*local_runtime.LocalPluginRuntime, int32)        {}
func (m *mockNotifier) OnLocalRuntimeInstanceLog(*local_runtime.LocalPluginRuntime, *local_runtime.PluginInstance, plugin_entities.PluginLogEvent) {
}
func (m *mockNotifier) OnDebuggingRuntimeConnected(r *debugging_runtime.RemotePluginRuntime) {
	m.connected = true
}
func (m *mockNotifier) OnDebuggingRuntimeDisconnected(r *debugging_runtime.RemotePluginRuntime) {
	m.disconnected = true
}

// setPrivateString sets an unexported string field on a struct value via unsafe reflection.
func setPrivateString(target any, field string, value string) {
	rv := reflect.ValueOf(target).Elem()
	f := rv.FieldByName(field)
	reflect.NewAt(f.Type(), unsafe.Pointer(f.UnsafeAddr())).Elem().SetString(value)
}

func newFakeRemoteRuntime(t *testing.T, name, version string) *debugging_runtime.RemotePluginRuntime {
	t.Helper()

	r := &debugging_runtime.RemotePluginRuntime{
		PluginRuntime: plugin_entities.PluginRuntime{
			Config: plugin_entities.PluginDeclaration{
				PluginDeclarationWithoutAdvancedFields: plugin_entities.PluginDeclarationWithoutAdvancedFields{
					// Author is overwritten with tenantId inside Identity()
					Name:    name,
					Version: manifest_entities.Version(version),
				},
			},
		},
	}

	// Provide required tenantId and checksum so Identity() succeeds
	setPrivateString(r, "tenantId", uuid.New().String())
	setPrivateString(r, "checksum", "0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef")
	reflect.NewAt(
		reflect.ValueOf(r).Elem().FieldByName("initialized").Type(),
		unsafe.Pointer(reflect.ValueOf(r).Elem().FieldByName("initialized").UnsafeAddr()),
	).Elem().SetBool(true)

	return r
}

func TestOnDebuggingRuntimeConnectedAndDisconnected(t *testing.T) {
	cp := NewControlPanel(&app.Config{PluginLocalLaunchingConcurrent: 1}, nil, nil, nil, nil)

	// Attach mock notifier to observe callbacks
	mn := &mockNotifier{}
	cp.AddNotifier(mn)

	r := newFakeRemoteRuntime(t, "conn_test", "1.2.3")

	// Connected path
	err := cp.onDebuggingRuntimeConnected(r)
	assert.NoError(t, err)

	id, err := r.Identity()
	assert.NoError(t, err)

	// Stored in runtime map and notifier called
	_, ok := cp.debuggingPluginRuntime.Load(id)
	assert.True(t, ok, "runtime should be stored on connect")
	assert.True(t, mn.connected, "connected notifier should be triggered")

	// Disconnected path
	cp.onDebuggingRuntimeDisconnected(r)

	_, ok = cp.debuggingPluginRuntime.Load(id)
	assert.False(t, ok, "runtime should be removed on disconnect")
	assert.True(t, mn.disconnected, "disconnected notifier should be triggered")
}

func TestOnDebuggingRuntimeDisconnected_IgnoresUninitializedRuntime(t *testing.T) {
	cp := NewControlPanel(&app.Config{PluginLocalLaunchingConcurrent: 1}, nil, nil, nil, nil)

	mn := &mockNotifier{}
	cp.AddNotifier(mn)

	r := &debugging_runtime.RemotePluginRuntime{}

	assert.NotPanics(t, func() {
		cp.onDebuggingRuntimeDisconnected(r)
	})
	assert.False(t, mn.disconnected, "disconnected notifier should not be triggered for uninitialized runtime")
}

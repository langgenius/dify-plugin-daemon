package plugin_manager

import (
	"reflect"
	"testing"
	"unsafe"

	"github.com/stretchr/testify/require"

	controlpanel "github.com/langgenius/dify-plugin-daemon/internal/core/control_panel"
	"github.com/langgenius/dify-plugin-daemon/internal/core/debugging_runtime"
	"github.com/langgenius/dify-plugin-daemon/internal/types/app"
	"github.com/langgenius/dify-plugin-daemon/pkg/entities/plugin_entities"
	"github.com/langgenius/dify-plugin-daemon/pkg/utils/mapping"
)

func TestGetPluginRuntimePrefersConnectedDebugRuntimeInServerlessMode(t *testing.T) {
	identifier := plugin_entities.PluginUniqueIdentifier(
		"langgenius/test_debug_plugin:0.0.1@0123456789abcdef0123456789abcdef",
	)
	debugRuntime := &debugging_runtime.RemotePluginRuntime{}
	controlPanel := controlpanel.NewControlPanel(
		&app.Config{PluginLocalLaunchingConcurrent: 1},
		nil,
		nil,
		nil,
		nil,
	)
	storeDebuggingPluginRuntime(t, controlPanel, identifier, debugRuntime)

	manager := &PluginManager{
		config:       &app.Config{Platform: app.PLATFORM_SERVERLESS},
		controlPanel: controlPanel,
	}

	runtime, err := manager.GetPluginRuntime(identifier)

	require.NoError(t, err)
	require.Same(t, debugRuntime, runtime)
}

func storeDebuggingPluginRuntime(
	t *testing.T,
	controlPanel *controlpanel.ControlPanel,
	identifier plugin_entities.PluginUniqueIdentifier,
	runtime *debugging_runtime.RemotePluginRuntime,
) {
	t.Helper()

	field := reflect.ValueOf(controlPanel).Elem().FieldByName("debuggingPluginRuntime")
	require.True(t, field.IsValid())

	runtimes := (*mapping.Map[
		plugin_entities.PluginUniqueIdentifier,
		*debugging_runtime.RemotePluginRuntime,
	])(unsafe.Pointer(field.UnsafeAddr()))
	runtimes.Store(identifier, runtime)
}

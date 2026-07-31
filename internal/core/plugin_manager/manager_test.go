package plugin_manager

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"

	controlpanel "github.com/langgenius/dify-plugin-daemon/internal/core/control_panel"
	"github.com/langgenius/dify-plugin-daemon/internal/core/debugging_runtime"
	invocationmock "github.com/langgenius/dify-plugin-daemon/internal/core/dify_invocation/mock"
	"github.com/langgenius/dify-plugin-daemon/internal/types/app"
	"github.com/langgenius/dify-plugin-daemon/pkg/entities/plugin_entities"
)

func TestBackwardsInvocationContextsAreIsolated(t *testing.T) {
	manager := &PluginManager{
		backwardsInvocation: invocationmock.NewMockedDifyInvocation(),
	}

	first := manager.BackwardsInvocation()
	second := manager.BackwardsInvocation()
	canceledContext, cancel := context.WithCancel(context.Background())
	first.SetContext(canceledContext)
	cancel()

	require.ErrorIs(t, first.Context().Err(), context.Canceled)
	require.NoError(t, second.Context().Err())
}

func TestNeedRedirectingUsesInstallationRuntimeType(t *testing.T) {
	identifier := plugin_entities.PluginUniqueIdentifier(
		"langgenius/test_debug_plugin:0.0.1@0123456789abcdef0123456789abcdef",
	)
	controlPanel := controlpanel.NewControlPanel(
		&app.Config{PluginLocalLaunchingConcurrent: 1},
		nil,
		nil,
		nil,
		nil,
	)
	manager := &PluginManager{
		config:       &app.Config{Platform: app.PLATFORM_SERVERLESS},
		controlPanel: controlPanel,
	}

	needRedirecting, err := manager.NeedRedirecting(
		identifier,
		plugin_entities.PLUGIN_RUNTIME_TYPE_REMOTE,
	)
	require.True(t, needRedirecting)
	require.ErrorIs(t, err, controlpanel.ErrPluginRuntimeNotFound)

	debugRuntime := &debugging_runtime.RemotePluginRuntime{}
	storeDebuggingPluginRuntime(t, controlPanel, identifier, debugRuntime)

	needRedirecting, err = manager.NeedRedirecting(
		identifier,
		plugin_entities.PLUGIN_RUNTIME_TYPE_REMOTE,
	)
	require.False(t, needRedirecting)
	require.NoError(t, err)
}

func TestNeedRedirectingDoesNotRedirectServerlessInstallation(t *testing.T) {
	identifier := plugin_entities.PluginUniqueIdentifier(
		"01234567-89ab-cdef-0123-456789abcdef/test_plugin:0.0.1@0123456789abcdef0123456789abcdef",
	)
	manager := &PluginManager{
		config: &app.Config{Platform: app.PLATFORM_SERVERLESS},
	}

	needRedirecting, err := manager.NeedRedirecting(
		identifier,
		plugin_entities.PLUGIN_RUNTIME_TYPE_SERVERLESS,
	)
	require.False(t, needRedirecting)
	require.NoError(t, err)
}

func TestNeedRedirectingFallsBackToIdentifierForDirectInvocation(t *testing.T) {
	identifier := plugin_entities.PluginUniqueIdentifier(
		"01234567-89ab-cdef-0123-456789abcdef/test_plugin:0.0.1@0123456789abcdef0123456789abcdef",
	)
	controlPanel := controlpanel.NewControlPanel(
		&app.Config{PluginLocalLaunchingConcurrent: 1},
		nil,
		nil,
		nil,
		nil,
	)
	manager := &PluginManager{
		config:       &app.Config{Platform: app.PLATFORM_SERVERLESS},
		controlPanel: controlPanel,
	}

	needRedirecting, err := manager.NeedRedirecting(identifier, "")
	require.True(t, needRedirecting)
	require.ErrorIs(t, err, controlpanel.ErrPluginRuntimeNotFound)
}

func TestNeedRedirectingKeepsDirectServerlessInvocationLocal(t *testing.T) {
	identifier := plugin_entities.PluginUniqueIdentifier(
		"langgenius/test_plugin:0.0.1@0123456789abcdef0123456789abcdef",
	)
	manager := &PluginManager{
		config: &app.Config{Platform: app.PLATFORM_SERVERLESS},
	}

	needRedirecting, err := manager.NeedRedirecting(identifier, "")
	require.False(t, needRedirecting)
	require.NoError(t, err)
}

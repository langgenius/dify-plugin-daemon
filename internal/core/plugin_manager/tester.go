package plugin_manager

import (
	"errors"
	"runtime"
	"time"

	"github.com/langgenius/dify-plugin-daemon/internal/core/plugin_daemon"
	"github.com/langgenius/dify-plugin-daemon/internal/core/plugin_daemon/access_types"
	"github.com/langgenius/dify-plugin-daemon/internal/core/plugin_manager/basic_manager"
	"github.com/langgenius/dify-plugin-daemon/internal/core/plugin_manager/local_manager"
	"github.com/langgenius/dify-plugin-daemon/internal/core/plugin_manager/positive_manager"
	"github.com/langgenius/dify-plugin-daemon/internal/core/session_manager"
	"github.com/langgenius/dify-plugin-daemon/internal/utils/log"
	"github.com/langgenius/dify-plugin-daemon/internal/utils/routine"
	"github.com/langgenius/dify-plugin-daemon/internal/utils/stream"
)

func (p *PluginManager) TestPlugin(
	path string,
	request map[string]any,
	access_type access_types.PluginAccessType,
	access_action access_types.PluginAccessAction,
	timeout string,
) (*stream.Stream[any], error) {
	// launch plugin runtime
	plugin, err := p.getLocalPluginRuntime(path)
	if err != nil {
		return nil, errors.Join(err, errors.New("failed to load plugin"))
	}

	// get assets
	assets, err := plugin.decoder.Assets()
	if err != nil {
		return nil, errors.Join(err, errors.New("failed to get assets"))
	}

	local_plugin_runtime := local_manager.NewLocalPluginRuntime(p.pythonInterpreterPath)
	local_plugin_runtime.PluginRuntime = plugin.runtime
	local_plugin_runtime.PositivePluginRuntime = positive_manager.PositivePluginRuntime{
		BasicPluginRuntime: basic_manager.NewBasicPluginRuntime(p.mediaManager),
		LocalPackagePath:   plugin.runtime.State.AbsolutePath,
		WorkingPath:        plugin.runtime.State.WorkingPath,
		Decoder:            plugin.decoder,
	}
	if err := local_plugin_runtime.RemapAssets(
		&local_plugin_runtime.Config,
		assets,
	); err != nil {
		return nil, errors.Join(err, errors.New("failed to remap assets"))
	}

	identity, err := local_plugin_runtime.Identity()
	if err != nil {
		return nil, errors.Join(err, errors.New("failed to get identity"))
	}

	// local plugin
	routine.Submit(func() {
		defer func() {
			if r := recover(); r != nil {
				// print stack trace
				buf := make([]byte, 1<<16)
				runtime.Stack(buf, true)
				log.Error("plugin runtime error: %v, stack trace: %s", r, string(buf))
			}
		}()
		// delete the plugin from the storage when the plugin is stopped
		p.fullDuplexLifetime(local_plugin_runtime, nil)
	})

	// wait for the plugin to start
	var timeout_duration time.Duration
	if timeout == "" {
		timeout_duration = 120 * time.Second
	} else {
		timeout_duration, err = time.ParseDuration(timeout)
		if err != nil {
			return nil, errors.Join(err, errors.New("failed to parse timeout"))
		}
	}
	select {
	case <-local_plugin_runtime.WaitStarted():
	case <-time.After(timeout_duration):
		return nil, errors.New("failed to start plugin after " + timeout_duration.String())
	}

	session := session_manager.NewSession(
		session_manager.NewSessionPayload{
			TenantID:               "test-tenant",
			UserID:                 "test-user",
			PluginUniqueIdentifier: identity,
			ClusterID:              "test-cluster",
			InvokeFrom:             access_type,
			Action:                 access_action,
			Declaration:            plugin.runtime.Configuration(),
			BackwardsInvocation:    manager.BackwardsInvocation(),
			IgnoreCache:            true,
		},
	)
	session.BindRuntime(local_plugin_runtime)
	defer session.Close(session_manager.CloseSessionPayload{
		IgnoreCache: true,
	})

	// try send request
	plugin_response, err := plugin_daemon.GenericInvokePlugin[map[string]any, any](session, &request, 1024)
	if err != nil {
		return nil, errors.Join(err, errors.New("failed to invoke plugin"))
	}

	plugin_response.OnClose(func() {
		local_plugin_runtime.Stop()
	})

	return plugin_response, nil
}

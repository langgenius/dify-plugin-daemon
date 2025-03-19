package plugin_manager

import (
	"errors"
	"fmt"
	"os"
	"path"
	"strings"

	"github.com/langgenius/dify-plugin-daemon/internal/core/plugin_manager/basic_runtime"
	"github.com/langgenius/dify-plugin-daemon/internal/core/plugin_manager/local_runtime"
	"github.com/langgenius/dify-plugin-daemon/internal/utils/log"
	"github.com/langgenius/dify-plugin-daemon/internal/utils/routine"
	"github.com/langgenius/dify-plugin-daemon/pkg/entities/plugin_entities"
	"github.com/langgenius/dify-plugin-daemon/pkg/plugin_packager/decoder"
)

type pluginRuntimeWithDecoder struct {
	runtime plugin_entities.PluginRuntime
	decoder decoder.PluginDecoder
}

// extract plugin from package to working directory
func (p *PluginManager) getLocalPluginRuntime(pluginUniqueIdentifier plugin_entities.PluginUniqueIdentifier) (
	*pluginRuntimeWithDecoder,
	error,
) {
	pluginZip, err := p.installedBucket.Get(pluginUniqueIdentifier)
	if err != nil {
		return nil, errors.Join(err, fmt.Errorf("get plugin package error"))
	}

	decoder, err := decoder.NewZipPluginDecoder(pluginZip)
	if err != nil {
		return nil, errors.Join(err, fmt.Errorf("create plugin decoder error"))
	}

	// get manifest
	manifest, err := decoder.Manifest()
	if err != nil {
		return nil, errors.Join(err, fmt.Errorf("get plugin manifest error"))
	}

	checksum, err := decoder.Checksum()
	if err != nil {
		return nil, errors.Join(err, fmt.Errorf("calculate checksum error"))
	}

	identity := manifest.Identity()
	identity = strings.ReplaceAll(identity, ":", "-")
	pluginWorkingPath := path.Join(p.workingDirectory, fmt.Sprintf("%s@%s", identity, checksum))
	return &pluginRuntimeWithDecoder{
		runtime: plugin_entities.PluginRuntime{
			Config: manifest,
			State: plugin_entities.PluginRuntimeState{
				Status:      plugin_entities.PLUGIN_RUNTIME_STATUS_PENDING,
				Restarts:    0,
				ActiveAt:    nil,
				Verified:    manifest.Verified,
				WorkingPath: pluginWorkingPath,
			},
		},
		decoder: decoder,
	}, nil
}

// launch a local plugin
// returns a full duplex lifetime, a launched channel, an error channel, and an error
// caller should always handle both the channels to avoid deadlock
// 1. for launched channel, launch process will close the channel to notify the caller, just wait for it
// 2. for error channel, it will be closed also, but no more error will be sent, caller should consume all errors
func (p *PluginManager) launchLocal(pluginUniqueIdentifier plugin_entities.PluginUniqueIdentifier) (
	plugin_entities.PluginFullDuplexLifetime, <-chan bool, <-chan error, error,
) {
	plugin, err := p.getLocalPluginRuntime(pluginUniqueIdentifier)
	if err != nil {
		return nil, nil, nil, err
	}

	identity, err := plugin.decoder.UniqueIdentity()
	if err != nil {
		return nil, nil, nil, err
	}

	// lock launch process
	p.localPluginLaunchingLock.Lock(identity.String())
	defer p.localPluginLaunchingLock.Unlock(identity.String())

	// check if the plugin is already running
	if lifetime, ok := p.m.Load(identity.String()); ok {
		lifetime, ok := lifetime.(plugin_entities.PluginFullDuplexLifetime)
		if !ok {
			return nil, nil, nil, fmt.Errorf("plugin runtime not found")
		}

		// returns a closed channel to indicate the plugin is already running, no more waiting is needed
		c := make(chan bool)
		close(c)
		errChan := make(chan error)
		close(errChan)

		return lifetime, c, errChan, nil
	}

	// extract plugin
	decoder, ok := plugin.decoder.(*decoder.ZipPluginDecoder)
	if !ok {
		return nil, nil, nil, fmt.Errorf("plugin decoder is not a zip decoder")
	}

	// check if the working directory exists, if not, create it, otherwise, launch it directly
	if _, err := os.Stat(plugin.runtime.State.WorkingPath); err != nil {
		if err := decoder.ExtractTo(plugin.runtime.State.WorkingPath); err != nil {
			return nil, nil, nil, errors.Join(err, fmt.Errorf("extract plugin to working directory error"))
		}
	}

	success := false
	failed := func(message string) error {
		if !success {
			os.RemoveAll(plugin.runtime.State.WorkingPath)
		}
		return errors.New(message)
	}

	// get assets
	assets, err := plugin.decoder.Assets()
	if err != nil {
		return nil, nil, nil, failed(err.Error())
	}

	localPluginRuntime := local_runtime.NewLocalPluginRuntime(local_runtime.LocalPluginRuntimeConfig{
		PythonInterpreterPath:     p.pythonInterpreterPath,
		PythonEnvInitTimeout:      p.pythonEnvInitTimeout,
		PythonCompileAllExtraArgs: p.pythonCompileAllExtraArgs,
		HttpProxy:                 p.HttpProxy,
		HttpsProxy:                p.HttpsProxy,
		PipMirrorUrl:              p.pipMirrorUrl,
		PipPreferBinary:           p.pipPreferBinary,
		PipExtraArgs:              p.pipExtraArgs,
	})
	localPluginRuntime.PluginRuntime = plugin.runtime
	localPluginRuntime.BasicChecksum = basic_runtime.BasicChecksum{
		MediaTransport: basic_runtime.NewMediaTransport(p.mediaBucket),
		WorkingPath:    plugin.runtime.State.WorkingPath,
		Decoder:        plugin.decoder,
	}

	if err := localPluginRuntime.RemapAssets(
		&localPluginRuntime.Config,
		assets,
	); err != nil {
		return nil, nil, nil, failed(errors.Join(err, fmt.Errorf("remap plugin assets error")).Error())
	}

	success = true

	p.m.Store(identity.String(), localPluginRuntime)

	// NOTE: you should always keep the size of the channel to 0
	// we use this to synchronize the plugin launch process
	launchedChan := make(chan bool)
	errChan := make(chan error)

	// local plugin
	routine.Submit(map[string]string{
		"module":   "plugin_manager",
		"function": "LaunchLocal",
	}, func() {
		defer func() {
			if r := recover(); r != nil {
				log.Error("plugin runtime panic: %v", r)
			}
			p.m.Delete(identity.String())
		}()

		// add max launching lock to prevent too many plugins launching at the same time
		p.maxLaunchingLock <- true
		routine.Submit(map[string]string{
			"module":   "plugin_manager",
			"function": "LaunchLocal",
		}, func() {
			// wait for plugin launched
			<-launchedChan
			// release max launching lock
			<-p.maxLaunchingLock
		})

		p.fullDuplexLifecycle(localPluginRuntime, launchedChan, errChan)
	})

	return localPluginRuntime, launchedChan, errChan, nil
}

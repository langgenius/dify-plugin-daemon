package plugin_manager

import (
	"errors"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path"
	"path/filepath"
	"strings"
	"time"

	"github.com/langgenius/dify-plugin-daemon/internal/core/plugin_manager/basic_manager"
	"github.com/langgenius/dify-plugin-daemon/internal/core/plugin_manager/local_manager"
	"github.com/langgenius/dify-plugin-daemon/internal/core/plugin_manager/positive_manager"
	"github.com/langgenius/dify-plugin-daemon/internal/core/plugin_manager/remote_manager"
	"github.com/langgenius/dify-plugin-daemon/internal/core/plugin_packager/decoder"
	"github.com/langgenius/dify-plugin-daemon/internal/types/app"
	"github.com/langgenius/dify-plugin-daemon/internal/types/entities/plugin_entities"
	"github.com/langgenius/dify-plugin-daemon/internal/utils/log"
	"github.com/langgenius/dify-plugin-daemon/internal/utils/routine"
)

func (p *PluginManager) startLocalWatcher() {
	go func() {
		log.Info("start to handle new plugins in path: %s", p.pluginStoragePath)
		p.handleNewLocalPlugins()
		for range time.NewTicker(time.Second * 30).C {
			p.handleNewLocalPlugins()
		}
	}()
}

func (p *PluginManager) initRemotePluginServer(config *app.Config) {
	if p.remotePluginServer != nil {
		return
	}
	p.remotePluginServer = remote_manager.NewRemotePluginServer(config, p.mediaManager)
}

func (p *PluginManager) startRemoteWatcher(config *app.Config) {
	// launch TCP debugging server if enabled
	if config.PluginRemoteInstallingEnabled {
		p.initRemotePluginServer(config)
		go func() {
			err := p.remotePluginServer.Launch()
			if err != nil {
				log.Error("start remote plugin server failed: %s", err.Error())
			}
		}()
		go func() {
			p.remotePluginServer.Wrap(func(rpr plugin_entities.PluginFullDuplexLifetime) {
				identity, err := rpr.Identity()
				if err != nil {
					log.Error("get remote plugin identity failed: %s", err.Error())
					return
				}
				p.m.Store(identity.String(), rpr)
				routine.Submit(func() {
					defer func() {
						if err := recover(); err != nil {
							log.Error("plugin runtime error: %v", err)
						}
						p.m.Delete(identity.String())
					}()
					p.fullDuplexLifecycle(rpr, nil)
				})
			})
		}()
	}
}

func (p *PluginManager) handleNewLocalPlugins() {
	// walk through all plugins
	err := filepath.WalkDir(p.pluginStoragePath, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if !d.IsDir() {
			_, _, err := p.launchLocal(path)
			if err != nil {
				log.Error("launch local plugin failed: %s", err.Error())
			}
		}

		return nil
	})

	if err != nil {
		log.Error("walk through plugins failed: %s", err.Error())
	}
}

func (p *PluginManager) launchLocal(plugin_package_path string) (
	plugin_entities.PluginFullDuplexLifetime, <-chan error, error,
) {
	plugin, err := p.getLocalPluginRuntime(plugin_package_path)
	if err != nil {
		return nil, nil, err
	}

	identity, err := plugin.decoder.UniqueIdentity()
	if err != nil {
		return nil, nil, err
	}

	// lock launch process
	p.localPluginLaunchingLock.Lock(identity.String())
	defer p.localPluginLaunchingLock.Unlock(identity.String())

	// check if the plugin is already running
	if lifetime, ok := p.m.Load(identity.String()); ok {
		lifetime, ok := lifetime.(plugin_entities.PluginFullDuplexLifetime)
		if !ok {
			return nil, nil, fmt.Errorf("plugin runtime not found")
		}

		// returns a closed channel to indicate the plugin is already running, no more waiting is needed
		c := make(chan error)
		close(c)

		return lifetime, c, nil
	}

	// extract plugin
	decoder, ok := plugin.decoder.(*decoder.ZipPluginDecoder)
	if !ok {
		return nil, nil, fmt.Errorf("plugin decoder is not a zip decoder")
	}

	// check if the working directory exists, if not, create it, otherwise, launch it directly
	if _, err := os.Stat(plugin.runtime.State.WorkingPath); err != nil {
		if err := decoder.ExtractTo(plugin.runtime.State.WorkingPath); err != nil {
			return nil, nil, errors.Join(err, fmt.Errorf("extract plugin to working directory error"))
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
		return nil, nil, failed(err.Error())
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
		return nil, nil, failed(errors.Join(err, fmt.Errorf("remap plugin assets error")).Error())
	}

	success = true

	p.m.Store(identity.String(), local_plugin_runtime)

	launched_chan := make(chan error)

	// local plugin
	routine.Submit(func() {
		defer func() {
			if r := recover(); r != nil {
				log.Error("plugin runtime panic: %v", r)
			}
			p.m.Delete(identity.String())
		}()

		// add max launching lock to prevent too many plugins launching at the same time
		p.maxLaunchingLock <- true
		routine.Submit(func() {
			// wait for plugin launched
			<-launched_chan
			// release max launching lock
			<-p.maxLaunchingLock
		})

		p.fullDuplexLifecycle(local_plugin_runtime, launched_chan)
	})

	return local_plugin_runtime, launched_chan, nil
}

type pluginRuntimeWithDecoder struct {
	runtime plugin_entities.PluginRuntime
	decoder decoder.PluginDecoder
}

// extract plugin from package to working directory
func (p *PluginManager) getLocalPluginRuntime(plugin_path string) (
	*pluginRuntimeWithDecoder,
	error,
) {
	pack, err := os.Open(plugin_path)
	if err != nil {
		return nil, errors.Join(err, fmt.Errorf("open plugin package error"))
	}
	defer pack.Close()

	if info, err := pack.Stat(); err != nil {
		return nil, errors.Join(err, fmt.Errorf("get plugin package info error"))
	} else if info.Size() > p.maxPluginPackageSize {
		log.Error("plugin package size is too large: %d", info.Size())
		return nil, errors.Join(err, fmt.Errorf("plugin package size is too large"))
	}

	plugin_zip, err := io.ReadAll(pack)
	if err != nil {
		return nil, errors.Join(err, fmt.Errorf("read plugin package error"))
	}

	decoder, err := decoder.NewZipPluginDecoder(plugin_zip)
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
	plugin_working_path := path.Join(p.workingDirectory, fmt.Sprintf("%s@%s", identity, checksum))
	return &pluginRuntimeWithDecoder{
		runtime: plugin_entities.PluginRuntime{
			Config: manifest,
			State: plugin_entities.PluginRuntimeState{
				Status:       plugin_entities.PLUGIN_RUNTIME_STATUS_PENDING,
				Restarts:     0,
				AbsolutePath: plugin_path,
				WorkingPath:  plugin_working_path,
				ActiveAt:     nil,
				Verified:     manifest.Verified,
			},
		},
		decoder: decoder,
	}, nil
}

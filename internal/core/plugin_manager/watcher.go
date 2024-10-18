package plugin_manager

import (
	"errors"
	"fmt"
	"io"
	"os"
	"path"
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

func (p *PluginManager) startRemoteWatcher(config *app.Config) {
	// launch TCP debugging server if enabled
	if config.PluginRemoteInstallingEnabled {
		server := remote_manager.NewRemotePluginServer(config, p.mediaManager)
		go func() {
			err := server.Launch()
			if err != nil {
				log.Error("start remote plugin server failed: %s", err.Error())
			}
		}()
		go func() {
			server.Wrap(func(rpr *remote_manager.RemotePluginRuntime) {
				routine.Submit(func() {
					defer func() {
						if err := recover(); err != nil {
							log.Error("plugin runtime error: %v", err)
						}
					}()
					p.fullDuplexLifetime(rpr)
				})
			})
		}()
	}
}

func (p *PluginManager) handleNewLocalPlugins() {
	// load local plugins firstly
	plugins, err := os.ReadDir(p.pluginStoragePath)
	if err != nil {
		log.Error("no plugin found in path: %s", p.pluginStoragePath)
	}

	for _, plugin := range plugins {
		if !plugin.IsDir() {
			abs_path := path.Join(p.pluginStoragePath, plugin.Name())
			_, err := p.launchLocal(abs_path)
			if err != nil {
				log.Error("launch local plugin failed: %s", err.Error())
			}
		}
	}
}

func (p *PluginManager) launchLocal(plugin_package_path string) (plugin_entities.PluginFullDuplexLifetime, error) {
	plugin, err := p.getLocalPluginRuntime(plugin_package_path)
	if err != nil {
		return nil, err
	}

	identity, err := plugin.decoder.UniqueIdentity()
	if err != nil {
		return nil, err
	}

	// lock launch process
	p.localPluginLaunchingLock.Lock(identity.String())
	defer p.localPluginLaunchingLock.Unlock(identity.String())

	// check if the plugin is already running
	if _, ok := p.m.Load(identity.String()); ok {
		lifetime, ok := p.Get(identity).(plugin_entities.PluginFullDuplexLifetime)
		if !ok {
			return nil, fmt.Errorf("plugin runtime not found")
		}
		return lifetime, nil
	}

	// extract plugin
	decoder, ok := plugin.decoder.(*decoder.ZipPluginDecoder)
	if !ok {
		return nil, fmt.Errorf("plugin decoder is not a zip decoder")
	}

	if err := decoder.ExtractTo(plugin.runtime.State.WorkingPath); err != nil {
		return nil, errors.Join(err, fmt.Errorf("extract plugin to working directory error"))
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
		return nil, failed(err.Error())
	}

	local_plugin_runtime := local_manager.NewLocalPluginRuntime()
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
		return nil, failed(errors.Join(err, fmt.Errorf("remap plugin assets error")).Error())
	}

	// add plugin to manager
	err = p.Add(local_plugin_runtime)
	if err != nil {
		return nil, failed(errors.Join(err, fmt.Errorf("add plugin to manager failed")).Error())
	}

	success = true

	// local plugin
	routine.Submit(func() {
		defer func() {
			if r := recover(); r != nil {
				log.Error("plugin runtime panic: %v", r)
			}
		}()
		p.fullDuplexLifetime(local_plugin_runtime)
	})

	return local_plugin_runtime, nil
}

type pluginRuntimeWithDecoder struct {
	runtime plugin_entities.PluginRuntime
	decoder decoder.PluginDecoder
}

// extract plugin from package to working directory
func (p *PluginManager) getLocalPluginRuntime(plugin_path string) (*pluginRuntimeWithDecoder, error) {
	pack, err := os.Open(plugin_path)
	if err != nil {
		return nil, errors.Join(err, fmt.Errorf("open plugin package error"))
	}
	defer pack.Close()

	if info, err := pack.Stat(); err != nil {
		return nil, errors.Join(err, fmt.Errorf("get plugin package info error"))
	} else if info.Size() > p.maxPluginPackageSize {
		log.Error("plugin package size is too large: %d", info.Size())
		return nil, err
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

	// check if already exists
	if _, exist := p.m.Load(manifest.Identity()); exist {
		return nil, errors.Join(fmt.Errorf("plugin already exists: %s", manifest.Identity()), err)
	}

	checksum, err := decoder.Checksum()
	if err != nil {
		return nil, errors.Join(err, fmt.Errorf("calculate checksum error"))
	}

	identity := manifest.Identity()
	// replace : with -
	identity = strings.ReplaceAll(identity, ":", "-")

	plugin_working_path := path.Join(p.workingDirectory, fmt.Sprintf("%s@%s", identity, checksum))

	// check if working directory exists
	if _, err := os.Stat(plugin_working_path); err == nil {
		return nil, errors.Join(fmt.Errorf("plugin working directory already exists: %s", plugin_working_path), err)
	}

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

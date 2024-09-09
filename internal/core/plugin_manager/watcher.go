package plugin_manager

import (
	"errors"
	"fmt"
	"io"
	"os"
	"path"
	"time"

	"github.com/langgenius/dify-plugin-daemon/internal/core/plugin_manager/basic_manager"
	"github.com/langgenius/dify-plugin-daemon/internal/core/plugin_manager/local_manager"
	"github.com/langgenius/dify-plugin-daemon/internal/core/plugin_manager/positive_manager"
	"github.com/langgenius/dify-plugin-daemon/internal/core/plugin_manager/remote_manager"
	"github.com/langgenius/dify-plugin-daemon/internal/core/plugin_packager/checksum"
	"github.com/langgenius/dify-plugin-daemon/internal/core/plugin_packager/decoder"
	"github.com/langgenius/dify-plugin-daemon/internal/core/plugin_packager/verifier"
	"github.com/langgenius/dify-plugin-daemon/internal/types/app"
	"github.com/langgenius/dify-plugin-daemon/internal/types/entities/plugin_entities"
	"github.com/langgenius/dify-plugin-daemon/internal/utils/log"
	"github.com/langgenius/dify-plugin-daemon/internal/utils/routine"
)

func (p *PluginManager) startLocalWatcher(config *app.Config) {
	go func() {
		log.Info("start to handle new plugins in path: %s", config.PluginStoragePath)
		p.handleNewLocalPlugins(config)
		for range time.NewTicker(time.Second * 30).C {
			p.handleNewLocalPlugins(config)
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
				p.localLifetime(rpr)
			})
		}()
	}
}

func (p *PluginManager) handleNewLocalPlugins(config *app.Config) {
	// load local plugins firstly
	for plugin := range p.loadNewLocalPlugins(config.PluginStoragePath) {
		// get assets
		assets, err := plugin.Decoder.Assets()
		if err != nil {
			log.Error("get plugin assets error: %v", err)
			continue
		}

		local_plugin_runtime := &local_manager.LocalPluginRuntime{
			PluginRuntime: plugin.Runtime,
			PositivePluginRuntime: positive_manager.PositivePluginRuntime{
				BasicPluginRuntime: basic_manager.NewBasicPluginRuntime(p.mediaManager),
				LocalPackagePath:   plugin.Runtime.State.AbsolutePath,
				WorkingPath:        plugin.Runtime.State.WorkingPath,
				Decoder:            plugin.Decoder,
			},
		}
		if err := local_plugin_runtime.RemapAssets(
			&local_plugin_runtime.Config,
			assets,
		); err != nil {
			log.Error("remap plugin assets error: %v", err)
			continue
		}

		identity, err := local_plugin_runtime.Identity()
		if err != nil {
			log.Error("get plugin identity error: %v", err)
			continue
		}

		// store the plugin in the storage, avoid duplicate loading
		p.runningPluginInStorage.Store(plugin.Runtime.State.AbsolutePath, identity.String())

		// local plugin
		routine.Submit(func() {
			defer func() {
				if r := recover(); r != nil {
					log.Error("plugin runtime error: %v", r)
				}
			}()
			// delete the plugin from the storage when the plugin is stopped
			defer p.runningPluginInStorage.Delete(plugin.Runtime.State.AbsolutePath)
			p.localLifetime(local_plugin_runtime)
		})
	}
}

type pluginRuntimeWithDecoder struct {
	Runtime plugin_entities.PluginRuntime
	Decoder decoder.PluginDecoder
}

// chan should be closed after using that
func (p *PluginManager) loadNewLocalPlugins(root_path string) <-chan *pluginRuntimeWithDecoder {
	ch := make(chan *pluginRuntimeWithDecoder)

	plugins, err := os.ReadDir(root_path)
	if err != nil {
		log.Error("no plugin found in path: %s", root_path)
		close(ch)
		return ch
	}

	routine.Submit(func() {
		for _, plugin := range plugins {
			if !plugin.IsDir() {
				abs_path := path.Join(root_path, plugin.Name())
				if _, ok := p.runningPluginInStorage.Load(abs_path); ok {
					// if the plugin is already running, skip it
					continue
				}

				plugin, err := p.loadPlugin(abs_path)
				if err != nil {
					log.Error("load plugin error: %v", err)
					continue
				}
				ch <- plugin
			}
		}
		close(ch)
	})

	return ch
}

func (p *PluginManager) loadPlugin(plugin_path string) (*pluginRuntimeWithDecoder, error) {
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

	checksum, err := checksum.CalculateChecksum(decoder)
	if err != nil {
		return nil, errors.Join(err, fmt.Errorf("calculate checksum error"))
	}

	plugin_working_path := path.Join(p.workingDirectory, fmt.Sprintf("%s@%s", manifest.Identity(), checksum))

	// check if working directory exists
	if _, err := os.Stat(plugin_working_path); err == nil {
		return nil, errors.Join(fmt.Errorf("plugin working directory already exists: %s", plugin_working_path), err)
	}

	// copy to working directory
	if err := decoder.Walk(func(filename, dir string) error {
		working_path := path.Join(plugin_working_path, dir)
		// check if directory exists
		if err := os.MkdirAll(working_path, 0755); err != nil {
			return err
		}

		bytes, err := decoder.ReadFile(filename)
		if err != nil {
			return err
		}

		filename = path.Join(working_path, filename)

		// copy file
		if err := os.WriteFile(filename, bytes, 0644); err != nil {
			return err
		}

		return nil
	}); err != nil {
		return nil, errors.Join(fmt.Errorf("copy plugin to working directory error: %v", err), err)
	}

	return &pluginRuntimeWithDecoder{
		Runtime: plugin_entities.PluginRuntime{
			Config: manifest,
			State: plugin_entities.PluginRuntimeState{
				Status:       plugin_entities.PLUGIN_RUNTIME_STATUS_PENDING,
				Restarts:     0,
				AbsolutePath: plugin_path,
				WorkingPath:  plugin_working_path,
				ActiveAt:     nil,
				Verified:     verifier.VerifyPlugin(decoder) == nil,
			},
		},
		Decoder: decoder,
	}, nil
}

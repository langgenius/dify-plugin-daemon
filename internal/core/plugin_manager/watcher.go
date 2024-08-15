package plugin_manager

import (
	"fmt"
	"io"
	"os"
	"path"
	"time"

	"github.com/langgenius/dify-plugin-daemon/internal/core/plugin_manager/aws_manager"
	"github.com/langgenius/dify-plugin-daemon/internal/core/plugin_manager/local_manager"
	"github.com/langgenius/dify-plugin-daemon/internal/core/plugin_manager/positive_manager"
	"github.com/langgenius/dify-plugin-daemon/internal/core/plugin_manager/remote_manager"
	"github.com/langgenius/dify-plugin-daemon/internal/core/plugin_packager/decoder"
	"github.com/langgenius/dify-plugin-daemon/internal/core/plugin_packager/verifier"
	"github.com/langgenius/dify-plugin-daemon/internal/types/app"
	"github.com/langgenius/dify-plugin-daemon/internal/types/entities"
	"github.com/langgenius/dify-plugin-daemon/internal/types/entities/plugin_entities"
	"github.com/langgenius/dify-plugin-daemon/internal/utils/log"
	"github.com/langgenius/dify-plugin-daemon/internal/utils/routine"
)

func (p *PluginManager) startLocalWatcher(config *app.Config) {
	go func() {
		log.Info("start to handle new plugins in path: %s", config.PluginStoragePath)
		p.handleNewPlugins(config)
		for range time.NewTicker(time.Second * 30).C {
			p.handleNewPlugins(config)
		}
	}()
}

func (p *PluginManager) startRemoteWatcher(config *app.Config) {
	// launch TCP debugging server if enabled
	if config.PluginRemoteInstallingEnabled {
		server := remote_manager.NewRemotePluginServer(config)
		go func() {
			err := server.Launch()
			if err != nil {
				log.Error("start remote plugin server failed: %s", err.Error())
			}
		}()
		go func() {
			server.Wrap(func(rpr *remote_manager.RemotePluginRuntime) {
				p.lifetime(config, rpr)
			})
		}()
	}
}

func (p *PluginManager) handleNewPlugins(config *app.Config) {
	// load local plugins firstly
	for plugin := range p.loadNewPlugins(config.PluginStoragePath) {
		var plugin_interface entities.PluginRuntimeInterface

		if config.Platform == app.PLATFORM_AWS_LAMBDA {
			plugin_interface = &aws_manager.AWSPluginRuntime{
				PluginRuntime: plugin.Runtime,
				PositivePluginRuntime: positive_manager.PositivePluginRuntime{
					LocalPackagePath: plugin.Runtime.State.AbsolutePath,
					WorkingPath:      plugin.Runtime.State.WorkingPath,
					Decoder:          plugin.Decoder,
				},
			}
		} else if config.Platform == app.PLATFORM_LOCAL {
			plugin_interface = &local_manager.LocalPluginRuntime{
				PluginRuntime: plugin.Runtime,
				PositivePluginRuntime: positive_manager.PositivePluginRuntime{
					LocalPackagePath: plugin.Runtime.State.AbsolutePath,
					WorkingPath:      plugin.Runtime.State.WorkingPath,
					Decoder:          plugin.Decoder,
				},
			}
		} else {
			log.Error("unsupported platform: %s for plugin: %s", config.Platform, plugin.Runtime.Config.Name)
			continue
		}

		routine.Submit(func() {
			p.lifetime(config, plugin_interface)
		})
	}
}

type pluginRuntimeWithDecoder struct {
	Runtime entities.PluginRuntime
	Decoder decoder.PluginDecoder
}

// chan should be closed after using that
func (p *PluginManager) loadNewPlugins(root_path string) <-chan *pluginRuntimeWithDecoder {
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
				plugin, err := p.loadPlugin(path.Join(root_path, plugin.Name()))
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
		log.Error("open plugin package error: %v", err)
		return nil, err
	}
	defer pack.Close()

	if info, err := pack.Stat(); err != nil {
		log.Error("get plugin package info error: %v", err)
		return nil, err
	} else if info.Size() > p.maxPluginPackageSize {
		log.Error("plugin package size is too large: %d", info.Size())
		return nil, err
	}

	plugin_zip, err := io.ReadAll(pack)
	if err != nil {
		log.Error("read plugin package error: %v", err)
		return nil, err
	}

	decoder, err := decoder.NewZipPluginDecoder(plugin_zip)
	if err != nil {
		log.Error("create plugin decoder error: %v", err)
		return nil, err
	}

	// get manifest
	manifest, err := decoder.Manifest()
	if err != nil {
		log.Error("get plugin manifest error: %v", err)
		return nil, err
	}

	// check if already exists
	if _, exist := p.m.Load(manifest.Identity()); exist {
		log.Warn("plugin already exists: %s", manifest.Identity())
		return nil, fmt.Errorf("plugin already exists: %s", manifest.Identity())
	}

	plugin_working_path := path.Join(p.workingDirectory, manifest.Identity())

	// check if working directory exists
	if _, err := os.Stat(plugin_working_path); err == nil {
		log.Warn("plugin working directory already exists: %s", plugin_working_path)
		return nil, fmt.Errorf("plugin working directory already exists: %s", plugin_working_path)
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
		log.Error("copy plugin to working directory error: %v", err)
		return nil, err
	}

	return &pluginRuntimeWithDecoder{
		Runtime: entities.PluginRuntime{
			Config: manifest,
			State: entities.PluginRuntimeState{
				Status:       entities.PLUGIN_RUNTIME_STATUS_PENDING,
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

func parsePluginConfig(configuration_path string) (*plugin_entities.PluginDeclaration, error) {
	text, err := os.ReadFile(configuration_path)
	if err != nil {
		return nil, err
	}

	result, err := plugin_entities.UnmarshalPluginDeclarationFromYaml(text)
	if err != nil {
		return nil, err
	}

	return result, nil
}

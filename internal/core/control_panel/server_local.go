package controlpanel

import (
	"errors"
	"fmt"
	"sync"
	"time"

	"github.com/langgenius/dify-plugin-daemon/internal/core/plugin_manager/local_runtime"
	"github.com/langgenius/dify-plugin-daemon/internal/utils/log"
	"github.com/langgenius/dify-plugin-daemon/internal/utils/routine"
	"github.com/langgenius/dify-plugin-daemon/pkg/entities/plugin_entities"
	"github.com/langgenius/dify-plugin-daemon/pkg/plugin_packager/decoder"
)

func (c *ControlPanel) StartLocalMonitor() {
	log.Info("start to handle new plugins in path: %s", c.config.PluginInstalledPath)
	log.Info("launch plugins with max concurrency: %d", c.config.PluginLocalLaunchingConcurrent)

	c.handleNewLocalPlugins()
	// sync every 30 seconds
	for range time.NewTicker(time.Second * 30).C {
		c.handleNewLocalPlugins()
	}
}

func (c *ControlPanel) handleNewLocalPlugins() {
	// TODO: handle new local plugins
	// walk through all plugins
	plugins, err := c.installedBucket.List()
	if err != nil {
		log.Error("list installed plugins failed: %s", err.Error())
		return
	}

	var wg sync.WaitGroup
	maxConcurrency := config.PluginLocalLaunchingConcurrent
	sem := make(chan struct{}, maxConcurrency)

	for _, plugin := range plugins {
		// TODO: optimize following codes
		_, exist := p.m.Load(plugin.String())
		if exist {
			continue
		}

		wg.Add(1)
		// Fix closure issue: create local variable copy
		currentPlugin := plugin
		routine.Submit(map[string]string{
			"module":   "plugin_manager",
			"function": "handleNewLocalPlugins",
		}, func() {
			// Acquire sem inside goroutine
			sem <- struct{}{}
			defer func() {
				if err := recover(); err != nil {
					log.Error("plugin launch runtime error: %v", err)
				}
				<-sem
				wg.Done()
			}()

			_, launchedChan, errChan, err := p.launchLocal(currentPlugin)
			if err != nil {
				log.Error("launch local plugin failed: %s", err.Error())
				return
			}

			// Handle error channel
			if errChan != nil {
				for err := range errChan {
					log.Error("plugin launch error: %s", err.Error())
				}
			}

			// Wait for plugin to complete startup
			if launchedChan != nil {
				<-launchedChan
			}
		})
	}

	// wait for all plugins to be launched
	wg.Wait()
}

func (c *ControlPanel) getLocalPluginRuntime(
	pluginUniqueIdentifier plugin_entities.PluginUniqueIdentifier,
) (*local_runtime.LocalPluginRuntime, error) {
	pluginZip, err := c.installedBucket.Get(pluginUniqueIdentifier)
	if err != nil {
		return nil, errors.Join(err, fmt.Errorf("get plugin package error"))
	}

	decoder, err := decoder.NewZipPluginDecoder(pluginZip)
	if err != nil {
		return nil, errors.Join(err, fmt.Errorf("create plugin decoder error"))
	}

	return local_runtime.ConstructPluginRuntime(c.config, decoder)
}

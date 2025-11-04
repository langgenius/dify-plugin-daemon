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

var (
	MAX_RETRY_COUNT = int32(15)

	RETRY_WAIT_INTERVAL_MAP = map[int32]time.Duration{
		0:               0 * time.Second,
		3:               30 * time.Second,
		8:               60 * time.Second,
		MAX_RETRY_COUNT: 240 * time.Second,
		// stop
	}
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

	for _, uniquePluginIdentifier := range plugins {
		// TODO: optimize following codes
		if c.localPluginRuntimes.Exists(uniquePluginIdentifier) {
			continue
		}

		// get the retry count
		retry, ok := c.localPluginFailsRecord.Load(uniquePluginIdentifier)
		if !ok {
			retry = LocalPluginFailsRecord{
				RetryCount:  0,
				LastTriedAt: time.Now(),
			}
		}

		if retry.RetryCount >= MAX_RETRY_COUNT {
			continue
		}

		waitTime := 0 * time.Second
		// calculate the wait time
		for c, v := range RETRY_WAIT_INTERVAL_MAP {
			if retry.RetryCount <= c {
				waitTime = v
				break
			}
		}

		// if the wait time is not 0, and the last failed at is not too long ago, skip it
		if waitTime > 0 && time.Since(retry.LastTriedAt) < waitTime {
			continue
		}

		wg.Add(1)
		routine.Submit(map[string]string{
			"module":   "plugin_manager",
			"function": "handleNewLocalPlugins",
		}, func() {
			defer wg.Done()
			_, ch, err := c.LaunchLocalPlugin(uniquePluginIdentifier)
			if err != nil {
				log.Error("launch local plugin failed: %s", err.Error())
				return
			}

			err = <-ch
			if err != nil {
				// record the failure
				c.localPluginFailsRecord.Store(uniquePluginIdentifier, LocalPluginFailsRecord{
					RetryCount:  retry.RetryCount + 1,
					LastTriedAt: time.Now(),
				})
			} else {
				// reset the failure record
				c.localPluginFailsRecord.Delete(uniquePluginIdentifier)
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

	// create a decoder to verify the plugin package
	decoder, err := decoder.NewZipPluginDecoderWithThirdPartySignatureVerificationConfig(
		pluginZip, &decoder.ThirdPartySignatureVerificationConfig{
			Enabled:        c.config.ThirdPartySignatureVerificationEnabled,
			PublicKeyPaths: c.config.ThirdPartySignatureVerificationPublicKeys,
		},
	)
	if err != nil {
		return nil, errors.Join(err, fmt.Errorf("create plugin decoder error"))
	}

	return local_runtime.ConstructPluginRuntime(c.config, decoder)
}

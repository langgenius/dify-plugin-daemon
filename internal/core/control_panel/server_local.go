package controlpanel

import (
	"sync"
	"time"

	"github.com/langgenius/dify-plugin-daemon/internal/core/local_runtime"
	"github.com/langgenius/dify-plugin-daemon/pkg/entities/plugin_entities"
	routinepkg "github.com/langgenius/dify-plugin-daemon/pkg/routine"
	"github.com/langgenius/dify-plugin-daemon/pkg/utils/log"
	"github.com/langgenius/dify-plugin-daemon/pkg/utils/routine"
)

func (c *ControlPanel) startLocalMonitor() {
	log.Info("start to handle new plugins in path: %s", c.config.PluginInstalledPath)
	log.Info("launch plugins with max concurrency: %d", c.config.PluginLocalLaunchingConcurrent)

	c.handleNewLocalPlugins()
	// sync every 30 seconds
	for range time.NewTicker(time.Second * 30).C {
		c.handleNewLocalPlugins()
	}
}

// continue check if a plugin was uninstalled
// AS plugin_daemon supports cluster mode
// installed plugins were stored in `c.installedBucket`
// it's a stateless across all plugin_daemon nodes
// a plugin may be uninstalled by other nodes
// to ensure all uninstalled plugin running in all nodes are stopped
func (c *ControlPanel) removeUnusedLocalPlugins() {
	for range time.NewTicker(time.Second * 30).C {
		// remove already uninstalled plugins
		c.localPluginRuntimes.Range(func(
			key plugin_entities.PluginUniqueIdentifier,
			value *local_runtime.LocalPluginRuntime,
		) bool {
			// remove plugin runtime
			if exists, err := c.installedBucket.Exists(key); err != nil {
				log.Error("check if plugin %s is installed failed: %s", key.String(), err.Error())
			} else if !exists {
				// Trigger a signal to stop a local plugin runtime
				if _, err := c.ShutdownLocalPluginGracefully(key); err != nil {
					log.Error("shutdown local plugin %s failed: %s", key.String(), err.Error())
				}
			}

			return true
		})
	}
}

// continue check if a new plugin was installed.
// the same as `removeUnusedLocalPlugins`, it's a cluster system,
// the installation of a plugin may be triggered by other nodes
// sync all the installed plugins in all nodes
func (c *ControlPanel) handleNewLocalPlugins() {
	// walk through all plugins
	plugins, err := c.installedBucket.List()
	if err != nil {
		log.Error("list installed plugins failed: %s", err.Error())
		return
	}

	var wg sync.WaitGroup

	for _, uniquePluginIdentifier := range plugins {
		// check if the plugin is in the ignore list
		if _, ok := c.localPluginWatchIgnoreList.Load(uniquePluginIdentifier); ok {
			// skip the plugin
			continue
		}

		// skip if the plugin is already launched
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

		waitTime := c.calculateWaitTime(retry.RetryCount)
		// if the wait time is not 0, and the last failed at is not too long ago, skip it
		if waitTime > 0 && time.Since(retry.LastTriedAt) < waitTime {
			continue
		}

		wg.Add(1)
		routine.Submit(routinepkg.Labels{
			routinepkg.RoutineLabelKeyModule: "plugin_manager",
			routinepkg.RoutineLabelKeyMethod: "handleNewLocalPlugins",
		}, func() {
			defer wg.Done()
			_, ch, err := c.LaunchLocalPlugin(uniquePluginIdentifier)
			if err != nil {
				log.Error("launch local plugin failed: %s, retried in %d seconds", err.Error(), waitTime)
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

// calculate the wait time for a plugin to be launched
// return 0 if the retry count is 0
func (c *ControlPanel) calculateWaitTime(
	retryCount int32,
) time.Duration {
	waitTime := 0 * time.Second
	// calculate the wait time
	for c, v := range RETRY_WAIT_INTERVAL_MAP {
		// find the best match retry count
		if retryCount >= c && v >= waitTime {
			waitTime = v
		}
	}

	return waitTime
}

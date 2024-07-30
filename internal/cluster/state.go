package cluster

import (
	"sync/atomic"

	"github.com/langgenius/dify-plugin-daemon/internal/types/entities"
	"github.com/langgenius/dify-plugin-daemon/internal/utils/log"
)

// RegisterPlugin registers a plugin to the cluster, and start to be scheduled
func (c *Cluster) RegisterPlugin(lifetime entities.PluginRuntimeTimeLifeInterface) error {
	identity, err := lifetime.Identity()
	if err != nil {
		return err
	}

	done := make(chan bool)
	closed := new(int32)
	close := func() {
		if atomic.CompareAndSwapInt32(closed, 0, 1) {
			close(done)
		}
	}

	lifetime.OnStop(func() {
		c.plugin_lock.Lock()
		delete(c.plugins, identity)
		c.plugin_lock.Unlock()
		close()
	})

	c.plugin_lock.Lock()
	if !lifetime.Stopped() {
		c.plugins[identity] = &PluginLifeTime{
			lifetime: lifetime,
		}
	} else {
		close()
	}
	c.plugin_lock.Unlock()

	log.Info("start to schedule plugin %s", identity)

	return nil
}

// SchedulePlugin schedules a plugin to the cluster
func (c *Cluster) schedulePlugins() error {
	return nil
}

// doPluginUpdate updates the plugin state and schedule the plugin
func (c *Cluster) doPluginStateUpdate(lifetime entities.PluginRuntimeTimeLifeInterface) error {
	return nil
}

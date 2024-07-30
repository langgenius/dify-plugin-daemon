package cluster

import (
	"github.com/langgenius/dify-plugin-daemon/internal/types/entities"
	"github.com/langgenius/dify-plugin-daemon/internal/utils/log"
)

type PluginLifeTime struct {
	lifetime entities.PluginRuntimeTimeLifeInterface
}

// RegisterPlugin registers a plugin to the cluster, and start to be scheduled
func (c *Cluster) RegisterPlugin(lifetime entities.PluginRuntimeTimeLifeInterface) error {
	identity, err := lifetime.Identity()
	if err != nil {
		return err
	}

	lifetime.OnStop(func() {
		c.plugin_lock.Lock()
		delete(c.plugins, identity)
		c.plugin_lock.Unlock()
	})

	c.plugin_lock.Lock()
	if !lifetime.Stopped() {
		c.plugins[identity] = &PluginLifeTime{
			lifetime: lifetime,
		}
	}
	c.plugin_lock.Unlock()

	log.Info("start to schedule plugin %s", identity)

	return nil
}

func (c *Cluster) SchedulePlugin(lifetime entities.PluginRuntimeTimeLifeInterface) error {
	return nil
}

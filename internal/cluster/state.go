package cluster

import (
	"sync/atomic"
	"time"

	"github.com/langgenius/dify-plugin-daemon/internal/types/entities"
	"github.com/langgenius/dify-plugin-daemon/internal/utils/cache"
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

	l := &pluginLifeTime{
		lifetime: lifetime,
	}

	lifetime.OnStop(func() {
		c.plugin_lock.Lock()
		c.plugins.Delete(identity)
		// remove plugin state
		c.doPluginStateUpdate(l)
		c.plugin_lock.Unlock()
		close()
	})

	c.plugin_lock.Lock()
	if !lifetime.Stopped() {
		c.plugins.Store(identity, l)

		// do plugin state update immediately
		err = c.doPluginStateUpdate(l)
		if err != nil {
			close()
			c.plugin_lock.Unlock()
			return err
		}
	} else {
		close()
	}
	c.plugin_lock.Unlock()

	log.Info("start to schedule plugin %s", identity)

	return nil
}

const (
	PLUGIN_STATE_MAP_KEY = "plugin_state"
)

func (c *Cluster) getPluginStateKey(node_id string, plugin_id string) string {
	return node_id + ":" + plugin_id
}

func (c *Cluster) getScanPluginsByNodeKey(node_id string) string {
	return node_id + ":*"
}

func (c *Cluster) getScanPluginsByIdKey(plugin_id string) string {
	return "*:" + plugin_id
}

func (c *Cluster) FetchPluginAvailableNodes(hashed_plugin_id string) ([]string, error) {
	states, err := cache.ScanMap[entities.PluginRuntimeState](PLUGIN_STATE_MAP_KEY, c.getScanPluginsByIdKey(hashed_plugin_id))
	if err != nil {
		return nil, err
	}

	nodes := make([]string, 0)
	for key := range states {
		// split key into node_id and plugin_id
		if len(key) < len(hashed_plugin_id)+1 {
			log.Error("unexpected plugin state key: %s", key)
			continue
		}
		node_id := key[:len(key)-len(hashed_plugin_id)-1]
		nodes = append(nodes, node_id)
	}

	return nodes, nil
}

// SchedulePlugin schedules a plugin to the cluster
// it will walk through the plugin state map and update all the states
// as for the plugin has exited, normally, it will be removed automatically
// but once a plugin is not removed, it will be gc by the master node
func (c *Cluster) schedulePlugins() error {
	c.plugins.Range(func(key string, value *pluginLifeTime) bool {
		// do plugin state update
		err := c.doPluginStateUpdate(value)
		if err != nil {
			log.Error("failed to update plugin state: %s", err.Error())
		}

		return true
	})

	return nil
}

// doPluginUpdate updates the plugin state and schedule the plugin
func (c *Cluster) doPluginStateUpdate(lifetime *pluginLifeTime) error {
	state := lifetime.lifetime.RuntimeState()
	hash_identity, err := lifetime.lifetime.HashedIdentity()
	if err != nil {
		return err
	}

	identity, err := lifetime.lifetime.Identity()
	if err != nil {
		return err
	}

	state_key := c.getPluginStateKey(c.id, hash_identity)

	// check if the plugin has been removed
	if !c.plugins.Exits(identity) {
		// remove state
		err = c.removePluginState(hash_identity)
		if err != nil {
			return err
		}
	} else {
		// update plugin state
		state.ScheduledAt = &[]time.Time{time.Now()}[0]
		lifetime.lifetime.UpdateState(state)
		err = cache.SetMapOneField(PLUGIN_STATE_MAP_KEY, state_key, state)
		if err != nil {
			return err
		}
	}

	return nil
}

func (c *Cluster) removePluginState(hashed_identity string) error {
	return cache.DelMapField(PLUGIN_STATE_MAP_KEY, c.getPluginStateKey(c.id, hashed_identity))
}

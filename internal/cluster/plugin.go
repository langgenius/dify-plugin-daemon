package cluster

import (
	"errors"
	"strings"
	"sync/atomic"
	"time"

	"github.com/langgenius/dify-plugin-daemon/internal/types/entities/plugin_entities"
	"github.com/langgenius/dify-plugin-daemon/internal/utils/cache"
	"github.com/langgenius/dify-plugin-daemon/internal/utils/log"
)

type pluginLifeTime struct {
	lifetime          plugin_entities.PluginRuntimeTimeLifeInterface
	last_scheduled_at time.Time
}

type pluginState struct {
	plugin_entities.PluginRuntimeState
	Identity string `json:"identity"`
}

// RegisterPlugin registers a plugin to the cluster, and start to be scheduled
func (c *Cluster) RegisterPlugin(lifetime plugin_entities.PluginRuntimeTimeLifeInterface) error {
	identity, err := lifetime.Identity()
	if err != nil {
		return err
	}

	if c.plugins.Exits(identity) {
		return errors.New("plugin has been registered")
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
	})

	c.plugin_lock.Lock()
	if !lifetime.Stopped() {
		c.plugins.Store(identity, l)

		// do plugin state update immediately
		err = c.doPluginStateUpdate(l)
		if err != nil {
			c.plugin_lock.Unlock()
			return err
		}
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

// SchedulePlugin schedules a plugin to the cluster
// it will walk through the plugin state map and update all the states
// as for the plugin has exited, normally, it will be removed automatically
// but once a plugin is not removed, it will be gc by the master node
func (c *Cluster) schedulePlugins() error {
	c.notifyPluginSchedule()
	defer c.notifyPluginScheduleCompleted()

	c.plugins.Range(func(key string, value *pluginLifeTime) bool {
		if time.Since(value.last_scheduled_at) < PLUGIN_SCHEDULER_INTERVAL {
			return true
		}
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

	schedule_state := &pluginState{
		Identity:           identity,
		PluginRuntimeState: state,
	}

	state_key := c.getPluginStateKey(c.id, hash_identity)

	// check if the plugin has been removed
	if !c.plugins.Exits(identity) {
		// remove state
		err = c.removePluginState(c.id, hash_identity)
		if err != nil {
			return err
		}
	} else {
		// update plugin state
		schedule_state.ScheduledAt = &[]time.Time{time.Now()}[0]
		err = cache.SetMapOneField(PLUGIN_STATE_MAP_KEY, state_key, schedule_state)
		if err != nil {
			return err
		}
		lifetime.lifetime.UpdateScheduledAt(*schedule_state.ScheduledAt)
	}

	lifetime.last_scheduled_at = time.Now()

	return nil
}

func (c *Cluster) removePluginState(node_id string, hashed_identity string) error {
	err := cache.DelMapField(PLUGIN_STATE_MAP_KEY, c.getPluginStateKey(node_id, hashed_identity))
	if err != nil {
		return err
	}

	log.Info("plugin %s has been removed from node %s", hashed_identity, c.id)

	return nil
}

// forceGCNodePlugins will force garbage collect all the plugins on the node
func (c *Cluster) forceGCNodePlugins(node_id string) error {
	return cache.ScanMapAsync[pluginState](
		PLUGIN_STATE_MAP_KEY,
		c.getScanPluginsByNodeKey(node_id),
		func(m map[string]pluginState) error {
			for _, plugin_state := range m {
				if err := c.forceGCNodePlugin(node_id, plugin_state.Identity); err != nil {
					return err
				}
			}
			return nil
		},
	)
}

// forceGCNodePlugin will force garbage collect the plugin on the node
func (c *Cluster) forceGCNodePlugin(node_id string, plugin_id string) error {
	if node_id == c.id {
		c.plugin_lock.Lock()
		c.plugins.Delete(plugin_id)
		c.plugin_lock.Unlock()
	}

	if err := c.removePluginState(node_id, plugin_entities.HashedIdentity(plugin_id)); err != nil {
		return err
	}

	return nil
}

// forceGCPluginByNodePluginJoin will force garbage collect the plugin by node_plugin_join
func (c *Cluster) forceGCPluginByNodePluginJoin(node_plugin_join string) error {
	return cache.DelMapField(PLUGIN_STATE_MAP_KEY, node_plugin_join)
}

func (c *Cluster) isPluginActive(state *pluginState) bool {
	return state != nil && state.ScheduledAt != nil && time.Since(*state.ScheduledAt) < 60*time.Second
}

func (c *Cluster) splitNodePluginJoin(node_plugin_join string) (node_id string, plugin_hashed_id string, err error) {
	split := strings.Split(node_plugin_join, ":")
	if len(split) != 2 {
		return "", "", errors.New("invalid node_plugin_join")
	}

	return split[0], split[1], nil
}

// autoGCPlugins will automatically garbage collect the plugins that are no longer active
func (c *Cluster) autoGCPlugins() error {
	// skip if already in auto gc
	if atomic.LoadInt32(&c.is_in_auto_gc_plugins) == 1 {
		return nil
	}
	defer atomic.StoreInt32(&c.is_in_auto_gc_plugins, 0)

	return cache.ScanMapAsync[pluginState](
		PLUGIN_STATE_MAP_KEY,
		"*",
		func(m map[string]pluginState) error {
			for node_plugin_join, plugin_state := range m {
				if !c.isPluginActive(&plugin_state) {
					node_id, _, err := c.splitNodePluginJoin(node_plugin_join)
					if err != nil {
						return err
					}

					// force gc the plugin
					if err := c.forceGCNodePlugin(node_id, plugin_state.Identity); err != nil {
						return err
					}

					// one more time to force gc the plugin, there is a possibility
					// that the hash value of plugin's identity is not the same as the node_plugin_join
					// so we need to force gc the plugin by node_plugin_join again
					if err := c.forceGCPluginByNodePluginJoin(node_plugin_join); err != nil {
						return err
					}
				}
			}
			return nil
		},
	)
}

func (c *Cluster) IsPluginNoCurrentNode(identity string) bool {
	_, ok := c.plugins.Load(identity)
	return ok
}

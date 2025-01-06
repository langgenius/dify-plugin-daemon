package cluster

import (
	"errors"
	"strings"
	"sync/atomic"
	"time"

	"github.com/langgenius/dify-plugin-daemon/internal/utils/cache"
	"github.com/langgenius/dify-plugin-daemon/internal/utils/log"
	"github.com/langgenius/dify-plugin-daemon/pkg/entities/plugin_entities"
)

type pluginLifeTime struct {
	lifetime        plugin_entities.PluginLifetime
	lastScheduledAt time.Time
}

type pluginState struct {
	plugin_entities.PluginRuntimeState
	Identity string `json:"identity"`
}

// RegisterPlugin registers a plugin to the cluster, and start to be scheduled
func (c *Cluster) RegisterPlugin(lifetime plugin_entities.PluginLifetime) error {
	identity, err := lifetime.Identity()
	if err != nil {
		return err
	}

	if c.showLog {
		log.Info("registering plugin %s", identity.String())
	}

	if c.plugins.Exists(identity.String()) {
		return errors.New("plugin has been registered")
	}

	l := &pluginLifeTime{
		lifetime: lifetime,
	}

	lifetime.OnStop(func() {
		c.pluginLock.Lock()
		c.plugins.Delete(identity.String())
		// remove plugin state
		c.doPluginStateUpdate(l)
		c.pluginLock.Unlock()
	})

	c.pluginLock.Lock()
	if !lifetime.Stopped() {
		c.plugins.Store(identity.String(), l)

		// do plugin state update immediately
		err = c.doPluginStateUpdate(l)
		if err != nil {
			c.pluginLock.Unlock()
			return err
		}
	}
	c.pluginLock.Unlock()

	if c.showLog {
		log.Info("start to schedule plugin %s", identity)
	}

	return nil
}

const (
	PLUGIN_STATE_MAP_KEY = "plugin_state"
)

func (c *Cluster) getPluginStateKey(nodeId string, plugin_id string) string {
	return nodeId + ":" + plugin_id
}

func (c *Cluster) getScanPluginsByNodeKey(nodeId string) string {
	return nodeId + ":*"
}

func (c *Cluster) getScanPluginsByIdKey(plugin_id string) string {
	return "*:" + plugin_id
}

// SchedulePlugin schedules a plugin to the cluster
// it will walk through the plugin state map and update all the states
// as for the plugin has exited, normally, it will be removed automatically
// but once a plugin is not removed, it will be gc by the master node
func (c *Cluster) schedulePlugins() error {
	if c.showLog {
		log.Info("scheduling %d plugins", c.plugins.Len())
	}

	c.notifyPluginSchedule()
	defer c.notifyPluginScheduleCompleted()

	c.plugins.Range(func(key string, value *pluginLifeTime) bool {
		if time.Since(value.lastScheduledAt) < c.pluginSchedulerInterval {
			return true
		}
		if c.showLog {
			log.Info("scheduling plugin %s", key)
		}
		// do plugin state update
		err := c.doPluginStateUpdate(value)
		if err != nil {
			log.Error("failed to update plugin state: %s", err.Error())
		}

		if c.showLog {
			log.Info("scheduled plugin %s", key)
		}

		return true
	})

	if c.showLog {
		log.Info("scheduled %d plugins", c.plugins.Len())
	}

	return nil
}

// doPluginUpdate updates the plugin state and schedule the plugin
func (c *Cluster) doPluginStateUpdate(lifetime *pluginLifeTime) error {
	state := lifetime.lifetime.RuntimeState()
	identity, err := lifetime.lifetime.Identity()
	if err != nil {
		return err
	}

	if c.showLog {
		log.Info("updating plugin state %s", identity.String())
	}

	hashedIdentity := plugin_entities.HashedIdentity(identity.String())

	scheduleState := &pluginState{
		Identity:           identity.String(),
		PluginRuntimeState: state,
	}

	stateKey := c.getPluginStateKey(c.id, hashedIdentity)

	// check if the plugin has been removed
	if !c.plugins.Exists(identity.String()) {
		if c.showLog {
			log.Info("removing plugin state %s due no longer exists", identity.String())
		}
		// remove state
		err = c.removePluginState(c.id, hashedIdentity)
		if err != nil {
			return err
		}
	} else {
		if c.showLog {
			log.Info("updating plugin state %s", identity.String())
		}
		// update plugin state
		scheduleState.ScheduledAt = &[]time.Time{time.Now()}[0]
		err = cache.SetMapOneField(PLUGIN_STATE_MAP_KEY, stateKey, scheduleState)
		if err != nil {
			return err
		}
		lifetime.lifetime.UpdateScheduledAt(*scheduleState.ScheduledAt)
		if c.showLog {
			log.Info("updated plugin state %s", identity.String())
		}
	}

	lifetime.lastScheduledAt = time.Now()

	return nil
}

func (c *Cluster) removePluginState(nodeId string, hashed_identity string) error {
	if c.showLog {
		log.Info("removing plugin state %s", hashed_identity)
	}
	err := cache.DelMapField(PLUGIN_STATE_MAP_KEY, c.getPluginStateKey(nodeId, hashed_identity))
	if err != nil {
		return err
	}

	if c.showLog {
		log.Info("plugin %s has been removed from node %s", hashed_identity, c.id)
	}

	return nil
}

// forceGCNodePlugins will force garbage collect all the plugins on the node
func (c *Cluster) forceGCNodePlugins(nodeId string) error {
	return cache.ScanMapAsync(
		PLUGIN_STATE_MAP_KEY,
		c.getScanPluginsByNodeKey(nodeId),
		func(m map[string]pluginState) error {
			for _, plugin_state := range m {
				if err := c.forceGCNodePlugin(nodeId, plugin_state.Identity); err != nil {
					return err
				}
			}
			return nil
		},
	)
}

// forceGCNodePlugin will force garbage collect the plugin on the node
func (c *Cluster) forceGCNodePlugin(nodeId string, plugin_id string) error {
	if nodeId == c.id {
		c.pluginLock.Lock()
		c.plugins.Delete(plugin_id)
		c.pluginLock.Unlock()
	}

	if err := c.removePluginState(nodeId, plugin_entities.HashedIdentity(plugin_id)); err != nil {
		return err
	}

	return nil
}

// forceGCPluginByNodePluginJoin will force garbage collect the plugin by node_plugin_join
func (c *Cluster) forceGCPluginByNodePluginJoin(node_plugin_join string) error {
	return cache.DelMapField(PLUGIN_STATE_MAP_KEY, node_plugin_join)
}

func (c *Cluster) isPluginActive(state *pluginState) bool {
	if state == nil {
		return false
	}
	if state.ScheduledAt == nil {
		return false
	}
	if time.Since(*state.ScheduledAt) > c.pluginDeactivatedTimeout {
		return false
	}
	return true
}

func (c *Cluster) splitNodePluginJoin(node_plugin_join string) (nodeId string, plugin_hashed_id string, err error) {
	split := strings.Split(node_plugin_join, ":")
	if len(split) != 2 {
		return "", "", errors.New("invalid node_plugin_join")
	}

	return split[0], split[1], nil
}

// autoGCPlugins will automatically garbage collect the plugins that are no longer active
func (c *Cluster) autoGCPlugins() error {
	// skip if already in auto gc
	if atomic.LoadInt32(&c.isInAutoGcPlugins) == 1 {
		return nil
	}
	defer atomic.StoreInt32(&c.isInAutoGcPlugins, 0)

	return cache.ScanMapAsync(
		PLUGIN_STATE_MAP_KEY,
		"*",
		func(m map[string]pluginState) error {
			for node_plugin_join, plugin_state := range m {
				if !c.isPluginActive(&plugin_state) {
					nodeId, _, err := c.splitNodePluginJoin(node_plugin_join)
					if err != nil {
						return err
					}

					// force gc the plugin
					if err := c.forceGCNodePlugin(nodeId, plugin_state.Identity); err != nil {
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

func (c *Cluster) IsPluginOnCurrentNode(identity plugin_entities.PluginUniqueIdentifier) (bool, error) {
	_, ok := c.plugins.Load(identity.String())
	if !ok {
		_, err := c.manager.Get(identity)
		if err != nil {
			return false, err
		} else {
			return true, nil
		}
	}
	return ok, nil
}

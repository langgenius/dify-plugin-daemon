package cluster

import (
	"strings"
	"time"

	"github.com/langgenius/dify-plugin-daemon/internal/utils/cache"
)

const (
	CLUSTER_STATE_TENANT_LOCK_PREFIX       = "cluster_state_tenant_lock"
	CLUSTER_STATE_PLUGIN_LOCK_PREFIX       = "cluster_state_plugin_lock"
	CLUSTER_UPDATE_NODE_STATUS_LOCK_PREFIX = "cluster_update_node_status_lock"
)

func (c *Cluster) LockTenant(tenant_id string) error {
	key := strings.Join([]string{CLUSTER_STATE_TENANT_LOCK_PREFIX, tenant_id}, ":")
	return cache.Lock(key, time.Second*5, time.Second)
}

func (c *Cluster) UnlockTenant(tenant_id string) error {
	key := strings.Join([]string{CLUSTER_STATE_TENANT_LOCK_PREFIX, tenant_id}, ":")
	return cache.Unlock(key)
}

func (c *Cluster) LockPlugin(plugin_id string) error {
	key := strings.Join([]string{CLUSTER_STATE_PLUGIN_LOCK_PREFIX, plugin_id}, ":")
	return cache.Lock(key, time.Second*5, time.Second)
}

func (c *Cluster) UnlockPlugin(plugin_id string) error {
	key := strings.Join([]string{CLUSTER_STATE_PLUGIN_LOCK_PREFIX, plugin_id}, ":")
	return cache.Unlock(key)
}

func (c *Cluster) LockNodeStatus(node_id string) error {
	key := strings.Join([]string{CLUSTER_UPDATE_NODE_STATUS_LOCK_PREFIX, node_id}, ":")
	return cache.Lock(key, time.Second*5, time.Second)
}

func (c *Cluster) UnlockNodeStatus(node_id string) error {
	key := strings.Join([]string{CLUSTER_UPDATE_NODE_STATUS_LOCK_PREFIX, node_id}, ":")
	return cache.Unlock(key)
}

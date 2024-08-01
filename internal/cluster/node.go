package cluster

import (
	"errors"
	"net"
	"strings"
	"sync/atomic"
	"time"

	"github.com/langgenius/dify-plugin-daemon/internal/types/entities"
	"github.com/langgenius/dify-plugin-daemon/internal/utils/cache"
	"github.com/langgenius/dify-plugin-daemon/internal/utils/log"
	"github.com/langgenius/dify-plugin-daemon/internal/utils/network"
	"github.com/langgenius/dify-plugin-daemon/internal/utils/parser"
)

// update the status of the node
func (c *Cluster) updateNodeStatus() error {
	if err := c.LockNodeStatus(c.id); err != nil {
		return err
	}
	defer c.UnlockNodeStatus(c.id)

	// update the status of the node
	node_status, err := cache.GetMapField[node](CLUSTER_STATUS_HASH_MAP_KEY, c.id)
	if err != nil {
		if err == cache.ErrNotFound {
			// try to get ips configs
			ips, err := network.FetchCurrentIps()
			if err != nil {
				return err
			}
			node_status = &node{
				Ips: parser.Map(func(from net.IP) ip {
					return ip{
						Address: from.String(),
						Votes:   []vote{},
					}
				}, ips),
			}
		} else {
			return err
		}
	} else {
		ips, err := network.FetchCurrentIps()
		if err != nil {
			return err
		}
		// add new ip if not exist
		for _, _ip := range ips {
			found := false
			for _, node_ip := range node_status.Ips {
				if node_ip.Address == _ip.String() {
					found = true
					break
				}
			}
			if !found {
				node_status.Ips = append(node_status.Ips, ip{
					Address: _ip.String(),
					Votes:   []vote{},
				})
			}
		}
	}

	// refresh the last ping time
	node_status.LastPingAt = time.Now().Unix()

	// update the status of the node
	if err := cache.SetMapOneField(CLUSTER_STATUS_HASH_MAP_KEY, c.id, node_status); err != nil {
		return err
	}

	// get all the nodes
	nodes, err := c.GetNodes()
	if err != nil {
		return err
	}

	// update self nodes map
	c.node_lock.Lock()
	defer c.node_lock.Unlock()

	c.nodes.Clear()
	for node_id, node := range nodes {
		c.nodes.Store(node_id, node)
	}

	return nil
}

func (c *Cluster) GetNodes() (map[string]node, error) {
	nodes, err := cache.GetMap[node](CLUSTER_STATUS_HASH_MAP_KEY)
	if err != nil {
		return nil, err
	}

	for node_id, node := range nodes {
		// filter out the disconnected nodes
		if !node.available() {
			delete(nodes, node_id)
		}
	}

	return nodes, nil
}

// FetchPluginAvailableNodes fetches the available nodes of the given plugin
func (c *Cluster) FetchPluginAvailableNodes(hashed_plugin_id string) ([]string, error) {
	states, err := cache.ScanMap[entities.PluginRuntimeState](
		PLUGIN_STATE_MAP_KEY, c.getScanPluginsByIdKey(hashed_plugin_id),
	)
	if err != nil {
		return nil, err
	}

	nodes := make([]string, 0)
	for key := range states {
		node_id, _, err := c.splitNodePluginJoin(key)
		if err != nil {
			continue
		}
		if c.nodes.Exits(node_id) {
			nodes = append(nodes, node_id)
		}
	}

	return nodes, nil
}

func (c *Cluster) IsMaster() bool {
	return c.i_am_master
}

func (c *Cluster) IsNodeAlive(node_id string) bool {
	node_status, err := cache.GetMapField[node](CLUSTER_STATUS_HASH_MAP_KEY, node_id)
	if err != nil {
		return false
	}

	return node_status.available()
}

// gc the nodes has already deactivated
func (c *Cluster) autoGCNodes() error {
	if atomic.LoadInt32(&c.is_in_auto_gc_nodes) == 1 {
		return nil
	}
	defer atomic.StoreInt32(&c.is_in_auto_gc_nodes, 0)

	var total_errors error
	add_error := func(err error) {
		if err != nil {
			if total_errors == nil {
				total_errors = err
			} else {
				total_errors = errors.Join(total_errors, err)
			}
		}
	}

	// get all nodes status
	nodes, err := cache.GetMap[node](CLUSTER_STATUS_HASH_MAP_KEY)
	if err == cache.ErrNotFound {
		return nil
	}

	for node_id, node_status := range nodes {
		// delete the node if it is disconnected
		if !node_status.available() {
			// gc the node
			if err := c.gcNode(node_id); err != nil {
				add_error(err)
				continue
			}
		}
	}

	return total_errors
}

// remove the resource associated with the node
func (c *Cluster) gcNode(node_id string) error {
	// remove all plugins associated with the node
	if err := c.forceGCNodePlugins(node_id); err != nil {
		return err
	}

	// remove the node from the cluster
	c.node_lock.Lock()
	c.nodes.Delete(node_id)
	c.node_lock.Unlock()

	if err := c.LockNodeStatus(node_id); err != nil {
		return err
	}
	defer c.UnlockNodeStatus(node_id)

	err := cache.DelMapField(CLUSTER_STATUS_HASH_MAP_KEY, node_id)
	if err != nil {
		return err
	} else {
		log.Info("node %s has been removed from the cluster due to being disconnected", node_id)
	}

	return nil
}

// remove self node from the cluster
func (c *Cluster) removeSelfNode() error {
	return c.gcNode(c.id)
}

const (
	CLUSTER_UPDATE_NODE_STATUS_LOCK_PREFIX = "cluster-update-node-status-lock"
)

func (c *Cluster) LockNodeStatus(node_id string) error {
	key := strings.Join([]string{CLUSTER_UPDATE_NODE_STATUS_LOCK_PREFIX, node_id}, ":")
	return cache.Lock(key, time.Second*5, time.Second)
}

func (c *Cluster) UnlockNodeStatus(node_id string) error {
	key := strings.Join([]string{CLUSTER_UPDATE_NODE_STATUS_LOCK_PREFIX, node_id}, ":")
	return cache.Unlock(key)
}

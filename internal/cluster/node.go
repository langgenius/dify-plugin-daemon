package cluster

import (
	"errors"
	"net"
	"strings"
	"sync/atomic"
	"time"

	"github.com/langgenius/dify-plugin-daemon/internal/utils/cache"
	"github.com/langgenius/dify-plugin-daemon/internal/utils/log"
	"github.com/langgenius/dify-plugin-daemon/internal/utils/network"
	"github.com/langgenius/dify-plugin-daemon/internal/utils/parser"
	"github.com/langgenius/dify-plugin-daemon/pkg/entities/plugin_entities"
)

// update the status of the node
func (c *Cluster) updateNodeStatus() error {
	c.notifyNodeUpdate()
	defer c.notifyNodeUpdateCompleted()

	if err := c.LockNodeStatus(c.id); err != nil {
		return err
	}
	defer c.UnlockNodeStatus(c.id)

	// update the status of the node
	nodeStatus, err := cache.GetMapField[node](CLUSTER_STATUS_HASH_MAP_KEY, c.id)
	if err != nil {
		if err == cache.ErrNotFound {
			// try to get ips configs
			ips, err := network.FetchCurrentIps()
			if err != nil {
				return err
			}
			nodeStatus = &node{
				Addresses: parser.Map(func(from net.IP) address {
					return address{
						Ip:    from.String(),
						Port:  c.port,
						Votes: []vote{},
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
			for _, node_ip := range nodeStatus.Addresses {
				if node_ip.Ip == _ip.String() {
					found = true
					break
				}
			}
			if !found {
				nodeStatus.Addresses = append(nodeStatus.Addresses, address{
					Ip:    _ip.String(),
					Port:  c.port,
					Votes: []vote{},
				})
			}
		}
	}

	// refresh the last ping time
	nodeStatus.LastPingAt = time.Now().Unix()

	// update the status of the node
	if err := cache.SetMapOneField(CLUSTER_STATUS_HASH_MAP_KEY, c.id, nodeStatus); err != nil {
		return err
	}

	// get all the nodes
	nodes, err := c.GetNodes()
	if err != nil {
		return err
	}

	// update self nodes map
	c.nodes.Clear()
	for nodeId, node := range nodes {
		c.nodes.Store(nodeId, node)
	}

	return nil
}

func (c *Cluster) isNodeAvailable(node *node) bool {
	return time.Since(time.Unix(node.LastPingAt, 0)) < c.nodeDisconnectedTimeout
}

func (c *Cluster) GetNodes() (map[string]node, error) {
	nodes, err := cache.GetMap[node](CLUSTER_STATUS_HASH_MAP_KEY)
	if err != nil {
		return nil, err
	}

	for nodeId, node := range nodes {
		// filter out the disconnected nodes
		if !c.isNodeAvailable(&node) {
			delete(nodes, nodeId)
		}
	}

	return nodes, nil
}

// FetchPluginAvailableNodesByHashedId fetches the available nodes of the given plugin
func (c *Cluster) FetchPluginAvailableNodesByHashedId(hashedPluginId string) ([]string, error) {
	states, err := cache.ScanMap[plugin_entities.PluginRuntimeState](
		PLUGIN_STATE_MAP_KEY, c.getScanPluginsByIdKey(hashedPluginId),
	)
	if err != nil {
		return nil, err
	}

	nodes := make([]string, 0)
	for key := range states {
		nodeId, _, err := c.splitNodePluginJoin(key)
		if err != nil {
			continue
		}
		if c.nodes.Exists(nodeId) {
			nodes = append(nodes, nodeId)
		}
	}

	return nodes, nil
}

func (c *Cluster) FetchPluginAvailableNodesById(plugin_id string) ([]string, error) {
	hashedPluginId := plugin_entities.HashedIdentity(plugin_id)
	return c.FetchPluginAvailableNodesByHashedId(hashedPluginId)
}

func (c *Cluster) IsMaster() bool {
	return c.iAmMaster
}

func (c *Cluster) IsNodeAlive(nodeId string) bool {
	nodeStatus, err := cache.GetMapField[node](CLUSTER_STATUS_HASH_MAP_KEY, nodeId)
	if err != nil {
		return false
	}

	return c.isNodeAvailable(nodeStatus)
}

// gc the nodes has already deactivated
func (c *Cluster) autoGCNodes() error {
	if atomic.LoadInt32(&c.isInAutoGcNodes) == 1 {
		return nil
	}
	defer atomic.StoreInt32(&c.isInAutoGcNodes, 0)

	var totalErrors error
	addError := func(err error) {
		if err != nil {
			if totalErrors == nil {
				totalErrors = err
			} else {
				totalErrors = errors.Join(totalErrors, err)
			}
		}
	}

	// get all nodes status
	nodes, err := cache.GetMap[node](CLUSTER_STATUS_HASH_MAP_KEY)
	if err == cache.ErrNotFound {
		return nil
	}

	for nodeId, nodeStatus := range nodes {
		// delete the node if it is disconnected
		if !c.isNodeAvailable(&nodeStatus) {
			// gc the node
			if err := c.gcNode(nodeId); err != nil {
				addError(err)
				continue
			}
		}
	}

	return totalErrors
}

// remove the resource associated with the node
func (c *Cluster) gcNode(nodeId string) error {
	// remove all plugins associated with the node
	if err := c.forceGCNodePlugins(nodeId); err != nil {
		return err
	}

	// remove the node from the cluster
	c.nodes.Delete(nodeId)

	if err := c.LockNodeStatus(nodeId); err != nil {
		return err
	}
	defer c.UnlockNodeStatus(nodeId)

	err := cache.DelMapField(CLUSTER_STATUS_HASH_MAP_KEY, nodeId)
	if err != nil {
		return err
	} else {
		log.Info("node %s has been removed from the cluster due to being disconnected", nodeId)
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

func (c *Cluster) LockNodeStatus(nodeId string) error {
	key := strings.Join([]string{CLUSTER_UPDATE_NODE_STATUS_LOCK_PREFIX, nodeId}, ":")
	return cache.Lock(key, time.Second*5, time.Second)
}

func (c *Cluster) UnlockNodeStatus(nodeId string) error {
	key := strings.Join([]string{CLUSTER_UPDATE_NODE_STATUS_LOCK_PREFIX, nodeId}, ":")
	return cache.Unlock(key)
}

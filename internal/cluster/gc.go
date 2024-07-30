package cluster

import (
	"errors"
	"time"

	"github.com/langgenius/dify-plugin-daemon/internal/utils/cache"
)

// gc the nodes has already deactivated
func (c *Cluster) gcNodes() error {
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
		if time.Since(time.Unix(node_status.LastPingAt, 0)) > NODE_DISCONNECTED_TIMEOUT {
			// gc the node
			if err := c.gcNode(node_id); err != nil {
				add_error(err)
				continue
			}

			// delete the node status
			if err := cache.DelMapField(CLUSTER_STATUS_HASH_MAP_KEY, node_id); err != nil {
				add_error(err)
			}
		}
	}

	return total_errors
}

// remove the resource associated with the node
func (c *Cluster) gcNode(node_id string) error {
	return nil
}

package cluster

import (
	"errors"
	"net"
	"time"

	"github.com/langgenius/dify-plugin-daemon/internal/cluster/cluster_id"
	"github.com/langgenius/dify-plugin-daemon/internal/utils/cache"
	"github.com/langgenius/dify-plugin-daemon/internal/utils/network"
	"github.com/langgenius/dify-plugin-daemon/internal/utils/parser"
)

// Plugin daemon will preemptively try to lock the slot to be the master of the cluster
// and keep update current status of the whole cluster
// once the master is no longer active, one of the slave will try to lock the slot again
// and become the new master
//
// Once a node becomes master, It will take responsibility to gc the nodes has already deactivated
// and all nodes should to maintenance their own status
//
// State:
//	- hashmap[cluster-status]
//		- node-id:
//			- list[ip]:
//				- address: string
//				- vote: int
//			- last_ping_at: int64
//	- preemption-lock: node-id
//	- node-status-upgrade-status
//
// A node will be removed from the cluster if it is no longer active

const (
	CLUSTER_STATUS_HASH_MAP_KEY = "cluster-status-hash-map"
	PREEMPTION_LOCK_KEY         = "cluster-master-preemption-lock"
)

// try lock the slot to be the master of the cluster
// returns:
//   - bool: true if the slot is locked by the node
//   - error: error if any
func (c *Cluster) lockMaster() (bool, error) {
	var final_error error

	for i := 0; i < 3; i++ {
		if success, err := cache.SetNX(PREEMPTION_LOCK_KEY, cluster_id.GetInstanceID(), MASTER_LOCK_EXPIRED_TIME); err != nil {
			// try again
			if final_error == nil {
				final_error = err
			} else {
				final_error = errors.Join(final_error, err)
			}
		} else if !success {
			return false, nil
		} else {
			return true, nil
		}
	}

	return false, final_error
}

// update master
func (c *Cluster) updateMaster() error {
	// update expired time of master key
	if _, err := cache.Expire(PREEMPTION_LOCK_KEY, MASTER_LOCK_EXPIRED_TIME); err != nil {
		return err
	}

	return nil
}

// update the status of the node
func (c *Cluster) updateNodeStatus() error {
	if err := c.LockNodeStatus(cluster_id.GetInstanceID()); err != nil {
		return err
	}
	defer c.UnlockNodeStatus(cluster_id.GetInstanceID())

	// update the status of the node
	node_status, err := cache.GetMapField[node](CLUSTER_STATUS_HASH_MAP_KEY, cluster_id.GetInstanceID())
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
	if err := cache.SetMapOneField(CLUSTER_STATUS_HASH_MAP_KEY, cluster_id.GetInstanceID(), node_status); err != nil {
		return err
	}

	return nil
}

func (c *Cluster) IsMaster() bool {
	return c.i_am_master
}

func (c *Cluster) IsNodeAlive(node_id string) bool {
	node_status, err := cache.GetMapField[node](CLUSTER_STATUS_HASH_MAP_KEY, node_id)
	if err != nil {
		return false
	}

	return time.Since(time.Unix(node_status.LastPingAt, 0)) < NODE_DISCONNECTED_TIMEOUT
}

package cluster

import (
	"errors"
	"net"
	"time"

	"github.com/langgenius/dify-plugin-daemon/internal/cluster/cluster_id"
	"github.com/langgenius/dify-plugin-daemon/internal/utils/cache"
	"github.com/langgenius/dify-plugin-daemon/internal/utils/log"
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

var (
	i_am_master = false
)

const (
	CLUSTER_STATUS_HASH_MAP_KEY = "cluster-status-hash-map"
	PREEMPTION_LOCK_KEY         = "cluster-master-preemption-lock"
)

const (
	MASTER_LOCKING_INTERVAL     = time.Millisecond * 500 // interval to try to lock the slot to be the master
	MASTER_LOCK_EXPIRED_TIME    = time.Second * 5        // expired time of master key
	MASTER_GC_INTERVAL          = time.Second * 10       // interval to do garbage collection of nodes has already deactivated
	NODE_VOTE_INTERVAL          = time.Second * 30       // interval to vote the ips of the nodes
	UPDATE_NODE_STATUS_INTERVAL = time.Second * 5        // interval to update the status of the node
	NODE_DISCONNECTED_TIMEOUT   = time.Second * 10       // once a node is no longer active, it will be removed from the cluster
)

// lifetime of the cluster
func (c *Cluster) clusterLifetime() {
	ticker_lock_master := time.NewTicker(MASTER_LOCKING_INTERVAL)
	defer ticker_lock_master.Stop()

	ticker_update_node_status := time.NewTicker(UPDATE_NODE_STATUS_INTERVAL)
	defer ticker_update_node_status.Stop()

	master_gc_ticker := time.NewTicker(MASTER_GC_INTERVAL)
	defer master_gc_ticker.Stop()

	node_vote_ticker := time.NewTicker(NODE_VOTE_INTERVAL)
	defer node_vote_ticker.Stop()

	if err := c.voteIps(); err != nil {
		log.Error("failed to vote the ips of the nodes: %s", err.Error())
	}

	for {
		select {
		case <-ticker_lock_master.C:
			if !i_am_master {
				// try lock the slot
				if success, err := c.lockMaster(); err != nil {
					log.Error("failed to lock the slot to be the master of the cluster: %s", err.Error())
				} else if success {
					i_am_master = true
					log.Info("current node has become the master of the cluster")
				} else {
					i_am_master = false
					log.Info("current node lost the master slot")
				}
			} else {
				// update the master
				if err := c.updateMaster(); err != nil {
					log.Error("failed to update the master: %s", err.Error())
				}
			}
		case <-ticker_update_node_status.C:
			if err := c.updateNodeStatus(); err != nil {
				log.Error("failed to update the status of the node: %s", err.Error())
			}
		case <-master_gc_ticker.C:
			if i_am_master {
				if err := c.gcNodes(); err != nil {
					log.Error("failed to gc the nodes has already deactivated: %s", err.Error())
				}
			}
		case <-node_vote_ticker.C:
			if err := c.voteIps(); err != nil {
				log.Error("failed to vote the ips of the nodes: %s", err.Error())
			}
		}
	}
}

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
	return i_am_master
}

func (c *Cluster) IsNodeAlive(node_id string) bool {
	node_status, err := cache.GetMapField[node](CLUSTER_STATUS_HASH_MAP_KEY, node_id)
	if err != nil {
		return false
	}

	return time.Since(time.Unix(node_status.LastPingAt, 0)) < NODE_DISCONNECTED_TIMEOUT
}

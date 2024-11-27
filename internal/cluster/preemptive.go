package cluster

import (
	"errors"

	"github.com/langgenius/dify-plugin-daemon/internal/utils/cache"
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
//		- node_id:
//			- list[ip]:
//				- address: string
//				- vote[]:
//					- node_id: string
//					- voted_at: int64
//					- failed: bool
//			- last_ping_at: int64
//	- preemption-lock: node_id
//

const (
	CLUSTER_STATUS_HASH_MAP_KEY = "cluster-nodes-status-hash-map"
	PREEMPTION_LOCK_KEY         = "cluster-master-preemption-lock"
)

// try lock the slot to be the master of the cluster
// returns:
//   - bool: true if the slot is locked by the node
//   - error: error if any
func (c *Cluster) lockMaster() (bool, error) {
	var finalError error

	for i := 0; i < 3; i++ {
		if success, err := cache.SetNX(PREEMPTION_LOCK_KEY, c.id, c.masterLockExpiredTime); err != nil {
			// try again
			if finalError == nil {
				finalError = err
			} else {
				finalError = errors.Join(finalError, err)
			}
		} else if !success {
			return false, nil
		} else {
			return true, nil
		}
	}

	return false, finalError
}

// update master
func (c *Cluster) updateMaster() error {
	// update expired time of master key
	if _, err := cache.Expire(PREEMPTION_LOCK_KEY, c.masterLockExpiredTime); err != nil {
		return err
	}

	return nil
}

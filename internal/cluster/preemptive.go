package cluster

import (
	"errors"
	"time"

	"github.com/langgenius/dify-plugin-daemon/pkg/utils/cache"
	"github.com/redis/go-redis/v9"
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

	const maxAttempts = 10
	for i := 0; i < maxAttempts; i++ {
		success, err := cache.SetNX(PREEMPTION_LOCK_KEY, c.id, c.masterLockExpiredTime)
		if err == nil {
			if !success {
				return false, nil
			}
			return true, nil
		}
		if redis.IsReadOnlyError(err) && i+1 < maxAttempts {
			time.Sleep(200 * time.Millisecond)
			continue
		}
		if finalError == nil {
			finalError = err
		} else {
			finalError = errors.Join(finalError, err)
		}
		if !redis.IsReadOnlyError(err) {
			break
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

package cluster

import (
	"time"

	"github.com/langgenius/dify-plugin-daemon/internal/utils/log"
)

const (
	MASTER_LOCKING_INTERVAL     = time.Millisecond * 500 // interval to try to lock the slot to be the master
	MASTER_LOCK_EXPIRED_TIME    = time.Second * 5        // expired time of master key
	MASTER_GC_INTERVAL          = time.Second * 10       // interval to do garbage collection of nodes has already deactivated
	NODE_VOTE_INTERVAL          = time.Second * 30       // interval to vote the ips of the nodes
	UPDATE_NODE_STATUS_INTERVAL = time.Second * 5        // interval to update the status of the node
	NODE_DISCONNECTED_TIMEOUT   = time.Second * 10       // once a node is no longer active, it will be removed from the cluster
	PLUGIN_SCHEDULER_INTERVAL   = time.Second * 3        // interval to schedule the plugins
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

	plugin_scheduler_ticker := time.NewTicker(PLUGIN_SCHEDULER_INTERVAL)
	defer plugin_scheduler_ticker.Stop()

	if err := c.voteIps(); err != nil {
		log.Error("failed to vote the ips of the nodes: %s", err.Error())
	}

	for {
		select {
		case <-ticker_lock_master.C:
			if !c.i_am_master {
				// try lock the slot
				if success, err := c.lockMaster(); err != nil {
					log.Error("failed to lock the slot to be the master of the cluster: %s", err.Error())
				} else if success {
					c.i_am_master = true
					log.Info("current node has become the master of the cluster")
				} else {
					c.i_am_master = false
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
			if c.i_am_master {
				if err := c.gcNodes(); err != nil {
					log.Error("failed to gc the nodes has already deactivated: %s", err.Error())
				}
			}
		case <-node_vote_ticker.C:
			if err := c.voteIps(); err != nil {
				log.Error("failed to vote the ips of the nodes: %s", err.Error())
			}
		case <-plugin_scheduler_ticker.C:
			if err := c.schedulePlugins(); err != nil {
				log.Error("failed to schedule the plugins: %s", err.Error())
			}
		case <-c.stop_chan:
			return
		}
	}
}

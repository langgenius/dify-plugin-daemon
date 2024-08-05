package cluster

import (
	"time"

	"github.com/langgenius/dify-plugin-daemon/internal/utils/cache"
	"github.com/langgenius/dify-plugin-daemon/internal/utils/log"
	"github.com/langgenius/dify-plugin-daemon/internal/utils/routine"
)

const (
	// master
	// the cluster master is responsible for managing garbage collection for both nodes and plugins.
	// typically, each node handles the garbage collection for its own plugins
	// However, if a node becomes inactive, the master takes over this task.
	// every node has an equal chance of becoming the master.
	// once a node is selected as the master, it is locked in that role.
	// If the master node becomes inactive, the master slot is released, allowing other nodes to attempt to take over the role.
	MASTER_LOCKING_INTERVAL  = time.Millisecond * 500 // interval to try to lock the slot to be the master
	MASTER_LOCK_EXPIRED_TIME = time.Second * 2        // expired time of master key
	MASTER_GC_INTERVAL       = time.Second * 10       // interval to do garbage collection of nodes has already deactivated

	// node
	// To determine the available IPs of the nodes, each node will vote for the IPs of other nodes.
	// this voting process will occur every $NODE_VOTE_INTERVAL.
	// simultaneously, all nodes will synchronize to the latest status in memory every $UPDATE_NODE_STATUS_INTERVAL.
	// each node will also update its own status to remain active. If a node becomes inactive, it will be removed from the cluster.
	NODE_VOTE_INTERVAL          = time.Second * 30 // interval to vote the ips of the nodes
	UPDATE_NODE_STATUS_INTERVAL = time.Second * 5  // interval to update the status of the node
	NODE_DISCONNECTED_TIMEOUT   = time.Second * 10 // once a node is no longer active, it will be removed from the cluster

	// plugin scheduler
	// each node will schedule its plugins every $PLUGIN_SCHEDULER_INTERVAL time
	// and schedule process will be triggered every $PLUGIN_SCHEDULER_TICKER_INTERVAL time
	// not all the plugins will be scheduled every time, only the plugins that are not scheduled in $PLUGIN_SCHEDULER_INTERVAL time will be scheduled
	// and the plugins that are not active will be removed from the cluster
	PLUGIN_SCHEDULER_TICKER_INTERVAL = time.Second * 3  // interval to schedule the plugins
	PLUGIN_SCHEDULER_INTERVAL        = time.Second * 10 // interval to schedule the plugins
	PLUGIN_DEACTIVATED_TIMEOUT       = time.Second * 30 // once a plugin is no longer active, it will be removed from the cluster
)

const (
	CLUSTER_NEW_NODE_CHANNEL = "cluster-new-node-channel"
)

// lifetime of the cluster
func (c *Cluster) clusterLifetime() {
	defer func() {
		if err := c.removeSelfNode(); err != nil {
			log.Error("failed to remove the self node from the cluster: %s", err.Error())
		}
		c.notifyClusterStopped()

		close(c.notify_cluster_stopped_chan)
		close(c.notify_become_master_chan)
		close(c.notify_master_gc_chan)
		close(c.notify_master_gc_completed_chan)
		close(c.notify_voting_chan)
		close(c.notify_voting_completed_chan)
		close(c.notify_plugin_schedule_chan)
		close(c.notify_plugin_schedule_completed_chan)
		close(c.notify_node_update_chan)
		close(c.notify_node_update_completed_chan)
	}()

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

	// vote for all ips and find the best one, prepare for later traffic scheduling
	routine.Submit(func() {
		if err := c.updateNodeStatus(); err != nil {
			log.Error("failed to update the status of the node: %s", err.Error())
		}

		if err := cache.Publish(CLUSTER_NEW_NODE_CHANNEL, newNodeEvent{
			NodeID: c.id,
		}); err != nil {
			log.Error("failed to publish the new node event: %s", err.Error())
		}

		if err := c.voteIps(); err != nil {
			log.Error("failed to vote the ips of the nodes: %s", err.Error())
		}
	})

	new_node_chan, cancel := cache.Subscribe[newNodeEvent](CLUSTER_NEW_NODE_CHANNEL)
	defer cancel()

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
					c.notifyBecomeMaster()
				} else {
					if c.i_am_master {
						c.i_am_master = false
						log.Info("current node has released the master slot")
					}
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
				c.notifyMasterGC()
				if err := c.autoGCNodes(); err != nil {
					log.Error("failed to gc the nodes have already deactivated: %s", err.Error())
				}
				if err := c.autoGCPlugins(); err != nil {
					log.Error("failed to gc the plugins have already stopped: %s", err.Error())
				}
				c.notifyMasterGCCompleted()
			}
		case <-node_vote_ticker.C:
			if err := c.voteIps(); err != nil {
				log.Error("failed to vote the ips of the nodes: %s", err.Error())
			}
		case _, ok := <-new_node_chan:
			if ok {
				// vote for the new node
				if err := c.voteIps(); err != nil {
					log.Error("failed to vote the ips of the nodes: %s", err.Error())
				}
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

package cluster

import (
	"sync"
	"sync/atomic"

	"github.com/google/uuid"
	"github.com/langgenius/dify-plugin-daemon/internal/types/app"
	"github.com/langgenius/dify-plugin-daemon/internal/utils/mapping"
)

type Cluster struct {
	// id is the unique id of the cluster
	id string

	// i_am_master is the flag to indicate whether the current node is the master node
	i_am_master bool

	// main http port of the current node
	port uint16

	// plugins stores all the plugin life time of the current node
	plugins     mapping.Map[string, *pluginLifeTime]
	plugin_lock sync.RWMutex

	// nodes stores all the nodes of the cluster
	nodes mapping.Map[string, node]

	// signals for waiting for the cluster to stop
	stop_chan chan bool
	stopped   int32

	is_in_auto_gc_nodes   int32
	is_in_auto_gc_plugins int32

	// channels to notify cluster event
	notify_become_master_chan             chan bool
	notify_master_gc_chan                 chan bool
	notify_master_gc_completed_chan       chan bool
	notify_voting_chan                    chan bool
	notify_voting_completed_chan          chan bool
	notify_plugin_schedule_chan           chan bool
	notify_plugin_schedule_completed_chan chan bool
	notify_node_update_chan               chan bool
	notify_node_update_completed_chan     chan bool
	notify_cluster_stopped_chan           chan bool
}

func NewCluster(config *app.Config) *Cluster {
	return &Cluster{
		id:        uuid.New().String(),
		port:      uint16(config.ServerPort),
		stop_chan: make(chan bool),

		notify_become_master_chan:             make(chan bool),
		notify_master_gc_chan:                 make(chan bool),
		notify_master_gc_completed_chan:       make(chan bool),
		notify_voting_chan:                    make(chan bool),
		notify_voting_completed_chan:          make(chan bool),
		notify_plugin_schedule_chan:           make(chan bool),
		notify_plugin_schedule_completed_chan: make(chan bool),
		notify_node_update_chan:               make(chan bool),
		notify_node_update_completed_chan:     make(chan bool),
		notify_cluster_stopped_chan:           make(chan bool),
	}
}

func (c *Cluster) Launch() {
	go c.clusterLifetime()
}

func (c *Cluster) Close() error {
	if atomic.CompareAndSwapInt32(&c.stopped, 0, 1) {
		close(c.stop_chan)
	}

	return nil
}

// trigger for master event
func (c *Cluster) notifyBecomeMaster() {
	if atomic.LoadInt32(&c.stopped) == 1 {
		return
	}

	select {
	case c.notify_become_master_chan <- true:
	default:
	}
}

// receive the master event
func (c *Cluster) NotifyBecomeMaster() <-chan bool {
	if atomic.LoadInt32(&c.stopped) == 1 {
		return nil
	}
	return c.notify_become_master_chan
}

// trigger for master gc event
func (c *Cluster) notifyMasterGC() {
	if atomic.LoadInt32(&c.stopped) == 1 {
		return
	}

	select {
	case c.notify_master_gc_chan <- true:
	default:
	}
}

// trigger for master gc completed event
func (c *Cluster) notifyMasterGCCompleted() {
	if atomic.LoadInt32(&c.stopped) == 1 {
		return
	}

	select {
	case c.notify_master_gc_completed_chan <- true:
	default:
	}
}

// trigger for voting event
func (c *Cluster) notifyVoting() {
	if atomic.LoadInt32(&c.stopped) == 1 {
		return
	}

	select {
	case c.notify_voting_chan <- true:
	default:
	}
}

// trigger for voting completed event
func (c *Cluster) notifyVotingCompleted() {
	if atomic.LoadInt32(&c.stopped) == 1 {
		return
	}

	select {
	case c.notify_voting_completed_chan <- true:
	default:
	}
}

// trigger for plugin schedule event
func (c *Cluster) notifyPluginSchedule() {
	if atomic.LoadInt32(&c.stopped) == 1 {
		return
	}

	select {
	case c.notify_plugin_schedule_chan <- true:
	default:
	}
}

// trigger for plugin schedule completed event
func (c *Cluster) notifyPluginScheduleCompleted() {
	if atomic.LoadInt32(&c.stopped) == 1 {
		return
	}

	select {
	case c.notify_plugin_schedule_completed_chan <- true:
	default:
	}
}

// trigger for node update event
func (c *Cluster) notifyNodeUpdate() {
	if atomic.LoadInt32(&c.stopped) == 1 {
		return
	}

	select {
	case c.notify_node_update_chan <- true:
	default:
	}
}

// trigger for node update completed event
func (c *Cluster) notifyNodeUpdateCompleted() {
	if atomic.LoadInt32(&c.stopped) == 1 {
		return
	}

	select {
	case c.notify_node_update_completed_chan <- true:
	default:
	}
}

// trigger for cluster stopped event
func (c *Cluster) notifyClusterStopped() {
	select {
	case c.notify_cluster_stopped_chan <- true:
	default:
	}
}

// receive the master gc event
func (c *Cluster) NotifyMasterGC() <-chan bool {
	if atomic.LoadInt32(&c.stopped) == 1 {
		return nil
	}
	return c.notify_master_gc_chan
}

// receive the master gc completed event
func (c *Cluster) NotifyMasterGCCompleted() <-chan bool {
	if atomic.LoadInt32(&c.stopped) == 1 {
		return nil
	}
	return c.notify_master_gc_completed_chan
}

// receive the voting event
func (c *Cluster) NotifyVoting() <-chan bool {
	if atomic.LoadInt32(&c.stopped) == 1 {
		return nil
	}
	return c.notify_voting_chan
}

// receive the voting completed event
func (c *Cluster) NotifyVotingCompleted() <-chan bool {
	if atomic.LoadInt32(&c.stopped) == 1 {
		return nil
	}
	return c.notify_voting_completed_chan
}

// receive the plugin schedule event
func (c *Cluster) NotifyPluginSchedule() <-chan bool {
	if atomic.LoadInt32(&c.stopped) == 1 {
		return nil
	}
	return c.notify_plugin_schedule_chan
}

// receive the plugin schedule completed event
func (c *Cluster) NotifyPluginScheduleCompleted() <-chan bool {
	if atomic.LoadInt32(&c.stopped) == 1 {
		return nil
	}
	return c.notify_plugin_schedule_completed_chan
}

// receive the node update event
func (c *Cluster) NotifyNodeUpdate() <-chan bool {
	if atomic.LoadInt32(&c.stopped) == 1 {
		return nil
	}
	return c.notify_node_update_chan
}

// receive the node update completed event
func (c *Cluster) NotifyNodeUpdateCompleted() <-chan bool {
	if atomic.LoadInt32(&c.stopped) == 1 {
		return nil
	}
	return c.notify_node_update_completed_chan
}

// receive the cluster stopped event
func (c *Cluster) NotifyClusterStopped() <-chan bool {
	return c.notify_cluster_stopped_chan
}

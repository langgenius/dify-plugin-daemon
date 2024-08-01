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

	// port is the health check port of the cluster
	port uint16

	// plugins stores all the plugin life time of the cluster
	plugins     mapping.Map[string, *pluginLifeTime]
	plugin_lock sync.RWMutex

	// nodes stores all the nodes of the cluster
	nodes     mapping.Map[string, node]
	node_lock sync.RWMutex

	// signals for waiting for the cluster to stop
	stop_chan chan bool
	stopped   *int32

	is_in_auto_gc_nodes   int32
	is_in_auto_gc_plugins int32
}

func NewCluster(config *app.Config) *Cluster {
	return &Cluster{
		id:        uuid.New().String(),
		port:      uint16(config.ServerPort),
		stop_chan: make(chan bool),
		stopped:   new(int32),
	}
}

func (c *Cluster) Launch() {
	go c.clusterLifetime()
}

func (c *Cluster) Close() error {
	if atomic.CompareAndSwapInt32(c.stopped, 0, 1) {
		close(c.stop_chan)
	}

	return nil
}

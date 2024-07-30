package cluster

import (
	"sync"
	"sync/atomic"
	"time"

	"github.com/langgenius/dify-plugin-daemon/internal/types/app"
	"github.com/langgenius/dify-plugin-daemon/internal/types/entities"
)

type pluginLifeTime struct {
	lifetime          entities.PluginRuntimeTimeLifeInterface
	last_scheduled_at time.Time
}

type Cluster struct {
	i_am_master bool

	port uint16

	plugins     map[string]*pluginLifeTime
	plugin_lock sync.Mutex

	stop_chan chan bool
	stopped   *int32
}

func NewCluster(config *app.Config) *Cluster {
	return &Cluster{
		port:      uint16(config.ServerPort),
		plugins:   make(map[string]*pluginLifeTime),
		stop_chan: make(chan bool),
		stopped:   new(int32),
	}
}

func (c *Cluster) Launch(config *app.Config) {
	go c.clusterLifetime()
}

func (c *Cluster) Close() {
	if atomic.CompareAndSwapInt32(c.stopped, 0, 1) {
		close(c.stop_chan)
	}
}

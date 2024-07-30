package cluster

import (
	"sync"

	"github.com/langgenius/dify-plugin-daemon/internal/types/app"
)

type Cluster struct {
	port uint16

	plugins     map[string]*PluginLifeTime
	plugin_lock sync.Mutex
}

func NewCluster(config *app.Config) *Cluster {
	return &Cluster{
		port:    uint16(config.ServerPort),
		plugins: make(map[string]*PluginLifeTime),
	}
}

func (c *Cluster) Launch(config *app.Config) {
	go c.clusterLifetime()
}

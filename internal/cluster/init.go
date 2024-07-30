package cluster

import "github.com/langgenius/dify-plugin-daemon/internal/types/app"

type Cluster struct {
	port uint16
}

var (
	cluster *Cluster
)

func Launch(config *app.Config) {
	cluster = &Cluster{
		port: uint16(config.ServerPort),
	}

	go func() {
		cluster.clusterLifetime()
	}()
}

func GetCluster() *Cluster {
	return cluster
}

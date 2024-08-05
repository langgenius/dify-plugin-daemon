package server

import (
	"github.com/langgenius/dify-plugin-daemon/internal/cluster"
)

type App struct {
	cluster *cluster.Cluster
}

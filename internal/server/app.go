package server

import (
	"github.com/langgenius/dify-plugin-daemon/internal/cluster"
	"github.com/langgenius/dify-plugin-daemon/internal/core/io_tunnel/backwards_invocation/transaction"
	"github.com/langgenius/dify-plugin-daemon/internal/core/plugin_manager"
)

type App struct {
	// cluster instance of this node
	// schedule all the tasks related to the cluster, like request direct
	cluster *cluster.Cluster

	// endpoint handler
	// customize behavior of endpoint
	endpointHandler EndpointHandler

	// serverless transaction handler
	// accept serverless transaction request and forward to the plugin daemon
	serverlessTransactionHandler *transaction.ServerlessTransactionHandler

	// plugin manager instance
	pluginManager *plugin_manager.PluginManager
}

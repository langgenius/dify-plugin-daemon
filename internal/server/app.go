package server

import (
	"github.com/langgenius/dify-plugin-daemon/internal/cluster"
	"github.com/langgenius/dify-plugin-daemon/internal/core/persistence"
	"github.com/langgenius/dify-plugin-daemon/internal/core/plugin_daemon/backwards_invocation/transaction"
)

type App struct {
	// cluster instance of this node
	// schedule all the tasks related to the cluster, like request direct
	cluster *cluster.Cluster

	// endpoint handler
	// customize behavior of endpoint
	endpoint_handler EndpointHandler

	// aws transaction handler
	// accept aws transaction request and forward to the plugin daemon
	aws_transaction_handler *transaction.AWSTransactionHandler

	// persistence
	persistence *persistence.Persistence
}

package server

import (
	"github.com/langgenius/dify-plugin-daemon/internal/cluster"
	"github.com/langgenius/dify-plugin-daemon/internal/core/plugin_daemon/backwards_invocation/transaction"
)

type App struct {
	// cluster instance of this node
	// schedule all the tasks related to the cluster, like request direct
	cluster *cluster.Cluster

	// webhook handler
	// customize behavior of endpoint
	webhook_handler WebhookHandler

	// aws transaction handler
	// accept aws transaction request and forward to the plugin daemon
	aws_transaction_handler *transaction.AWSTransactionHandler
}

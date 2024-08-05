package plugin_daemon

import (
	"github.com/langgenius/dify-plugin-daemon/internal/core/session_manager"
	"github.com/langgenius/dify-plugin-daemon/internal/types/entities/requests"
	"github.com/langgenius/dify-plugin-daemon/internal/utils/stream"
)

func InvokeWebhook(
	session *session_manager.Session,
	request *requests.RequestInvokeWebhook,
) (
	*stream.StreamResponse[[]byte], error,
) {
	// TODO: implement this function
	return nil, nil
}

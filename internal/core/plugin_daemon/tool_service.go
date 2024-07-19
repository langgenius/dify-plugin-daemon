package plugin_daemon

import (
	"github.com/langgenius/dify-plugin-daemon/internal/core/session_manager"
	"github.com/langgenius/dify-plugin-daemon/internal/types/entities/plugin_entities"
	"github.com/langgenius/dify-plugin-daemon/internal/types/entities/requests"
	"github.com/langgenius/dify-plugin-daemon/internal/utils/stream"
)

func InvokeTool(
	session *session_manager.Session,
	request *requests.RequestInvokeTool,
) (
	*stream.StreamResponse[plugin_entities.ToolResponseChunk], error,
) {
	return genericInvokePlugin[requests.RequestInvokeTool, plugin_entities.ToolResponseChunk](
		session,
		request,
		128,
		PLUGIN_ACCESS_TYPE_TOOL,
		PLUGIN_ACCESS_ACTION_INVOKE_TOOL,
	)
}

package plugin_daemon

import (
	"github.com/langgenius/dify-plugin-daemon/internal/core/plugin_daemon/backwards_invocation"
	"github.com/langgenius/dify-plugin-daemon/internal/core/session_manager"
	"github.com/langgenius/dify-plugin-daemon/internal/types/entities/requests"
	"github.com/langgenius/dify-plugin-daemon/internal/types/entities/tool_entities"
	"github.com/langgenius/dify-plugin-daemon/internal/utils/stream"
)

func InvokeTool(
	session *session_manager.Session,
	request *requests.RequestInvokeTool,
) (
	*stream.StreamResponse[tool_entities.ToolResponseChunk], error,
) {
	return genericInvokePlugin[requests.RequestInvokeTool, tool_entities.ToolResponseChunk](
		session,
		request,
		128,
		backwards_invocation.PLUGIN_ACCESS_TYPE_TOOL,
		backwards_invocation.PLUGIN_ACCESS_ACTION_INVOKE_TOOL,
	)
}

func ValidateToolCredentials(
	session *session_manager.Session,
	request *requests.RequestValidateToolCredentials,
) (
	*stream.StreamResponse[tool_entities.ValidateCredentialsResult], error,
) {
	return genericInvokePlugin[requests.RequestValidateToolCredentials, tool_entities.ValidateCredentialsResult](
		session,
		request,
		1,
		backwards_invocation.PLUGIN_ACCESS_TYPE_TOOL,
		backwards_invocation.PLUGIN_ACCESS_ACTION_VALIDATE_TOOL_CREDENTIALS,
	)
}

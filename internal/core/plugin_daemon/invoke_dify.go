package plugin_daemon

import (
	"fmt"

	"github.com/langgenius/dify-plugin-daemon/internal/core/dify_invocation"
	"github.com/langgenius/dify-plugin-daemon/internal/core/plugin_daemon/backwards_invocation"
	"github.com/langgenius/dify-plugin-daemon/internal/core/session_manager"
	"github.com/langgenius/dify-plugin-daemon/internal/types/entities"
	"github.com/langgenius/dify-plugin-daemon/internal/utils/parser"
)

func invokeDify(
	runtime entities.PluginRuntimeInterface,
	invoke_from PluginAccessType,
	session *session_manager.Session, data []byte,
) error {
	// unmarshal invoke data
	request, err := parser.UnmarshalJsonBytes[map[string]any](data)

	if err != nil {
		return fmt.Errorf("unmarshal invoke request failed: %s", err.Error())
	}

	// prepare invocation arguments
	request_handle, err := prepareDifyInvocationArguments(session, request)
	if err != nil {
		return err
	}
	defer request_handle.End()

	if invoke_from == PLUGIN_ACCESS_TYPE_MODEL {
		request_handle.WriteError(fmt.Errorf("you can not invoke dify from %s", invoke_from))
		return nil
	}

	// dispatch invocation task
	dispatchDifyInvocationTask(request_handle)

	return nil
}

func prepareDifyInvocationArguments(session *session_manager.Session, request map[string]any) (*backwards_invocation.BackwardsInvocation, error) {
	typ, ok := request["type"].(string)
	if !ok {
		return nil, fmt.Errorf("invoke request missing type: %s", request)
	}

	// get request id
	backwards_request_id, ok := request["backwards_request_id"].(string)
	if !ok {
		return nil, fmt.Errorf("invoke request missing request_id: %s", request)
	}

	// get request
	detailed_request, ok := request["request"].(map[string]any)
	if !ok {
		return nil, fmt.Errorf("invoke request missing request: %s", request)
	}

	return backwards_invocation.NewBackwardsInvocation(
		backwards_invocation.BackwardsInvocationType(typ),
		backwards_request_id, session, detailed_request,
	), nil
}

func dispatchDifyInvocationTask(handle *backwards_invocation.BackwardsInvocation) {
	switch handle.Type() {
	case dify_invocation.INVOKE_TYPE_TOOL:
		_, err := parser.MapToStruct[dify_invocation.InvokeToolRequest](handle.RequestData())
		if err != nil {
			handle.WriteError(fmt.Errorf("unmarshal invoke tool request failed: %s", err.Error()))
			return
		}

	default:
		handle.WriteError(fmt.Errorf("unsupported invoke type: %s", handle.Type()))
	}
}

func setTaskContext(session *session_manager.Session, r *dify_invocation.BaseInvokeDifyRequest) {
	r.TenantId = session.TenantID()
	r.UserId = session.UserID()
}

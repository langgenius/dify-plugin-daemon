package plugin_daemon

import (
	"fmt"

	"github.com/langgenius/dify-plugin-daemon/internal/core/dify_invocation"
	"github.com/langgenius/dify-plugin-daemon/internal/core/plugin_daemon/backwards_invocation"
	"github.com/langgenius/dify-plugin-daemon/internal/core/session_manager"
	"github.com/langgenius/dify-plugin-daemon/internal/types/entities"
	"github.com/langgenius/dify-plugin-daemon/internal/utils/parser"
	"github.com/langgenius/dify-plugin-daemon/internal/utils/routine"
)

func invokeDify(
	runtime entities.PluginRuntimeInterface,
	invoke_from PluginAccessType,
	session *session_manager.Session, data []byte,
) error {
	// unmarshal invoke data
	request, err := parser.UnmarshalJsonBytes2Map(data)
	if err != nil {
		return fmt.Errorf("unmarshal invoke request failed: %s", err.Error())
	}

	if request == nil {
		return fmt.Errorf("invoke request is empty")
	}

	// prepare invocation arguments
	request_handle, err := prepareDifyInvocationArguments(session, request)
	if err != nil {
		return err
	}

	if invoke_from == PLUGIN_ACCESS_TYPE_MODEL {
		request_handle.WriteError(fmt.Errorf("you can not invoke dify from %s", invoke_from))
		request_handle.EndResponse()
		return nil
	}

	// dispatch invocation task
	routine.Submit(func() {
		dispatchDifyInvocationTask(request_handle)
		defer request_handle.EndResponse()
	})

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
	request_data := handle.RequestData()
	tenant_id, err := handle.TenantID()
	if err != nil {
		handle.WriteError(fmt.Errorf("get tenant id failed: %s", err.Error()))
		return
	}
	request_data["tenant_id"] = tenant_id
	user_id, err := handle.UserID()
	if err != nil {
		handle.WriteError(fmt.Errorf("get user id failed: %s", err.Error()))
		return
	}
	request_data["user_id"] = user_id

	switch handle.Type() {
	case dify_invocation.INVOKE_TYPE_TOOL:
		r, err := parser.MapToStruct[dify_invocation.InvokeToolRequest](handle.RequestData())
		if err != nil {
			handle.WriteError(fmt.Errorf("unmarshal invoke tool request failed: %s", err.Error()))
			return
		}
		executeDifyInvocationToolTask(handle, r)
	default:
		handle.WriteError(fmt.Errorf("unsupported invoke type: %s", handle.Type()))
	}
}

func executeDifyInvocationToolTask(
	handle *backwards_invocation.BackwardsInvocation,
	request *dify_invocation.InvokeToolRequest,
) {
	response, err := dify_invocation.InvokeTool(request)
	if err != nil {
		handle.WriteError(fmt.Errorf("invoke tool failed: %s", err.Error()))
		return
	}

	for response.Next() {
		data, err := response.Read()
		if err != nil {
			return
		}

		handle.WriteResponse("stream", data)
	}
}

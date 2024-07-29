package plugin_daemon

import (
	"fmt"

	"github.com/langgenius/dify-plugin-daemon/internal/core/dify_invocation"
	"github.com/langgenius/dify-plugin-daemon/internal/core/plugin_daemon/backwards_invocation"
	"github.com/langgenius/dify-plugin-daemon/internal/core/session_manager"
	"github.com/langgenius/dify-plugin-daemon/internal/types/entities"
	"github.com/langgenius/dify-plugin-daemon/internal/types/entities/tool_entities"
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

var (
	dispatchMapping = map[dify_invocation.InvokeType]func(handle *backwards_invocation.BackwardsInvocation){
		dify_invocation.INVOKE_TYPE_TOOL: func(handle *backwards_invocation.BackwardsInvocation) {
			genericDispatchTask[dify_invocation.InvokeToolRequest](handle, executeDifyInvocationToolTask)
		},
		dify_invocation.INVOKE_TYPE_LLM: func(handle *backwards_invocation.BackwardsInvocation) {
			genericDispatchTask[dify_invocation.InvokeLLMRequest](handle, executeDifyInvocationLLMTask)
		},
		dify_invocation.INVOKE_TYPE_TEXT_EMBEDDING: func(handle *backwards_invocation.BackwardsInvocation) {
			genericDispatchTask[dify_invocation.InvokeTextEmbeddingRequest](handle, executeDifyInvocationTextEmbeddingTask)
		},
		dify_invocation.INVOKE_TYPE_RERANK: func(handle *backwards_invocation.BackwardsInvocation) {
			genericDispatchTask[dify_invocation.InvokeRerankRequest](handle, executeDifyInvocationRerankTask)
		},
		dify_invocation.INVOKE_TYPE_TTS: func(handle *backwards_invocation.BackwardsInvocation) {
			genericDispatchTask[dify_invocation.InvokeTTSRequest](handle, executeDifyInvocationTTSTask)
		},
		dify_invocation.INVOKE_TYPE_SPEECH2TEXT: func(handle *backwards_invocation.BackwardsInvocation) {
			genericDispatchTask[dify_invocation.InvokeSpeech2TextRequest](handle, executeDifyInvocationSpeech2TextTask)
		},
		dify_invocation.INVOKE_TYPE_MODERATION: func(handle *backwards_invocation.BackwardsInvocation) {
			genericDispatchTask[dify_invocation.InvokeModerationRequest](handle, executeDifyInvocationModerationTask)
		},
	}
)

func genericDispatchTask[T any](
	handle *backwards_invocation.BackwardsInvocation,
	dispatch func(
		handle *backwards_invocation.BackwardsInvocation,
		request *T,
	),
) {
	r, err := parser.MapToStruct[T](handle.RequestData())
	if err != nil {
		handle.WriteError(fmt.Errorf("unmarshal invoke tool request failed: %s", err.Error()))
		return
	}
	dispatch(handle, r)
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

	for t, v := range dispatchMapping {
		if t == handle.Type() {
			v(handle)
			return
		}
	}

	handle.WriteError(fmt.Errorf("unsupported invoke type: %s", handle.Type()))
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

	response.Wrap(func(t tool_entities.ToolResponseChunk) {
		handle.WriteResponse("stream", t)
	})
}

func executeDifyInvocationLLMTask(
	handle *backwards_invocation.BackwardsInvocation,
	request *dify_invocation.InvokeLLMRequest,
) {

}

func executeDifyInvocationTextEmbeddingTask(
	handle *backwards_invocation.BackwardsInvocation,
	request *dify_invocation.InvokeTextEmbeddingRequest,
) {

}

func executeDifyInvocationRerankTask(
	handle *backwards_invocation.BackwardsInvocation,
	request *dify_invocation.InvokeRerankRequest,
) {

}

func executeDifyInvocationTTSTask(
	handle *backwards_invocation.BackwardsInvocation,
	request *dify_invocation.InvokeTTSRequest,
) {

}

func executeDifyInvocationSpeech2TextTask(
	handle *backwards_invocation.BackwardsInvocation,
	request *dify_invocation.InvokeSpeech2TextRequest,
) {

}

func executeDifyInvocationModerationTask(
	handle *backwards_invocation.BackwardsInvocation,
	request *dify_invocation.InvokeModerationRequest,
) {

}

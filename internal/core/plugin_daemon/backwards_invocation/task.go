package backwards_invocation

import (
	"fmt"

	"github.com/langgenius/dify-plugin-daemon/internal/core/dify_invocation"
	"github.com/langgenius/dify-plugin-daemon/internal/core/session_manager"
	"github.com/langgenius/dify-plugin-daemon/internal/types/entities"
	"github.com/langgenius/dify-plugin-daemon/internal/types/entities/model_entities"
	"github.com/langgenius/dify-plugin-daemon/internal/types/entities/tool_entities"
	"github.com/langgenius/dify-plugin-daemon/internal/utils/parser"
	"github.com/langgenius/dify-plugin-daemon/internal/utils/routine"
)

func InvokeDify(
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

	// check permission
	if err := checkPermission(runtime, request_handle); err != nil {
		request_handle.WriteError(err)
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

var (
	permissionMapping = map[dify_invocation.InvokeType]map[string]any{
		dify_invocation.INVOKE_TYPE_TOOL: {
			"func": func(runtime entities.PluginRuntimeInterface) bool {
				return runtime.Configuration().Resource.Permission.AllowInvokeTool()
			},
			"error": "permission denied, you need to enable tool access in plugin manifest",
		},
		dify_invocation.INVOKE_TYPE_LLM: {
			"func": func(runtime entities.PluginRuntimeInterface) bool {
				return runtime.Configuration().Resource.Permission.AllowInvokeLLM()
			},
			"error": "permission denied, you need to enable llm access in plugin manifest",
		},
		dify_invocation.INVOKE_TYPE_TEXT_EMBEDDING: {
			"func": func(runtime entities.PluginRuntimeInterface) bool {
				return runtime.Configuration().Resource.Permission.AllowInvokeTextEmbedding()
			},
			"error": "permission denied, you need to enable text-embedding access in plugin manifest",
		},
		dify_invocation.INVOKE_TYPE_RERANK: {
			"func": func(runtime entities.PluginRuntimeInterface) bool {
				return runtime.Configuration().Resource.Permission.AllowInvokeRerank()
			},
			"error": "permission denied, you need to enable rerank access in plugin manifest",
		},
		dify_invocation.INVOKE_TYPE_TTS: {
			"func": func(runtime entities.PluginRuntimeInterface) bool {
				return runtime.Configuration().Resource.Permission.AllowInvokeTTS()
			},
			"error": "permission denied, you need to enable tts access in plugin manifest",
		},
		dify_invocation.INVOKE_TYPE_SPEECH2TEXT: {
			"func": func(runtime entities.PluginRuntimeInterface) bool {
				return runtime.Configuration().Resource.Permission.AllowInvokeSpeech2Text()
			},
			"error": "permission denied, you need to enable speech2text access in plugin manifest",
		},
		dify_invocation.INVOKE_TYPE_MODERATION: {
			"func": func(runtime entities.PluginRuntimeInterface) bool {
				return runtime.Configuration().Resource.Permission.AllowInvokeModeration()
			},
			"error": "permission denied, you need to enable moderation access in plugin manifest",
		},
		dify_invocation.INVOKE_TYPE_NODE: {
			"func": func(runtime entities.PluginRuntimeInterface) bool {
				return runtime.Configuration().Resource.Permission.AllowInvokeNode()
			},
			"error": "permission denied, you need to enable node access in plugin manifest",
		},
	}
)

func checkPermission(runtime entities.PluginRuntimeInterface, request_handle *BackwardsInvocation) error {
	permission, ok := permissionMapping[request_handle.Type()]
	if !ok {
		return fmt.Errorf("unsupported invoke type: %s", request_handle.Type())
	}

	permission_func, ok := permission["func"].(func(runtime entities.PluginRuntimeInterface) bool)
	if !ok {
		return fmt.Errorf("permission function not found: %s", request_handle.Type())
	}

	if !permission_func(runtime) {
		return fmt.Errorf(permission["error"].(string))
	}

	return nil
}

func prepareDifyInvocationArguments(session *session_manager.Session, request map[string]any) (*BackwardsInvocation, error) {
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

	return NewBackwardsInvocation(
		BackwardsInvocationType(typ),
		backwards_request_id, session, detailed_request,
	), nil
}

var (
	dispatchMapping = map[dify_invocation.InvokeType]func(handle *BackwardsInvocation){
		dify_invocation.INVOKE_TYPE_TOOL: func(handle *BackwardsInvocation) {
			genericDispatchTask[dify_invocation.InvokeToolRequest](handle, executeDifyInvocationToolTask)
		},
		dify_invocation.INVOKE_TYPE_LLM: func(handle *BackwardsInvocation) {
			genericDispatchTask[dify_invocation.InvokeLLMRequest](handle, executeDifyInvocationLLMTask)
		},
		dify_invocation.INVOKE_TYPE_TEXT_EMBEDDING: func(handle *BackwardsInvocation) {
			genericDispatchTask[dify_invocation.InvokeTextEmbeddingRequest](handle, executeDifyInvocationTextEmbeddingTask)
		},
		dify_invocation.INVOKE_TYPE_RERANK: func(handle *BackwardsInvocation) {
			genericDispatchTask[dify_invocation.InvokeRerankRequest](handle, executeDifyInvocationRerankTask)
		},
		dify_invocation.INVOKE_TYPE_TTS: func(handle *BackwardsInvocation) {
			genericDispatchTask[dify_invocation.InvokeTTSRequest](handle, executeDifyInvocationTTSTask)
		},
		dify_invocation.INVOKE_TYPE_SPEECH2TEXT: func(handle *BackwardsInvocation) {
			genericDispatchTask[dify_invocation.InvokeSpeech2TextRequest](handle, executeDifyInvocationSpeech2TextTask)
		},
		dify_invocation.INVOKE_TYPE_MODERATION: func(handle *BackwardsInvocation) {
			genericDispatchTask[dify_invocation.InvokeModerationRequest](handle, executeDifyInvocationModerationTask)
		},
	}
)

func genericDispatchTask[T any](
	handle *BackwardsInvocation,
	dispatch func(
		handle *BackwardsInvocation,
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

func dispatchDifyInvocationTask(handle *BackwardsInvocation) {
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
	typ := handle.Type()
	request_data["type"] = typ

	for t, v := range dispatchMapping {
		if t == handle.Type() {
			v(handle)
			return
		}
	}

	handle.WriteError(fmt.Errorf("unsupported invoke type: %s", handle.Type()))
}

func executeDifyInvocationToolTask(
	handle *BackwardsInvocation,
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
	handle *BackwardsInvocation,
	request *dify_invocation.InvokeLLMRequest,
) {
	response, err := dify_invocation.InvokeLLM(request)
	if err != nil {
		handle.WriteError(fmt.Errorf("invoke llm model failed: %s", err.Error()))
		return
	}

	response.Wrap(func(t model_entities.LLMResultChunk) {
		handle.WriteResponse("stream", t)
	})
}

func executeDifyInvocationTextEmbeddingTask(
	handle *BackwardsInvocation,
	request *dify_invocation.InvokeTextEmbeddingRequest,
) {
	response, err := dify_invocation.InvokeTextEmbedding(request)
	if err != nil {
		handle.WriteError(fmt.Errorf("invoke text-embedding model failed: %s", err.Error()))
		return
	}

	handle.WriteResponse("struct", response)
}

func executeDifyInvocationRerankTask(
	handle *BackwardsInvocation,
	request *dify_invocation.InvokeRerankRequest,
) {
	response, err := dify_invocation.InvokeRerank(request)
	if err != nil {
		handle.WriteError(fmt.Errorf("invoke rerank model failed: %s", err.Error()))
		return
	}

	handle.WriteResponse("struct", response)
}

func executeDifyInvocationTTSTask(
	handle *BackwardsInvocation,
	request *dify_invocation.InvokeTTSRequest,
) {
	response, err := dify_invocation.InvokeTTS(request)
	if err != nil {
		handle.WriteError(fmt.Errorf("invoke tts model failed: %s", err.Error()))
		return
	}

	response.Wrap(func(t model_entities.TTSResult) {
		handle.WriteResponse("struct", t)
	})
}

func executeDifyInvocationSpeech2TextTask(
	handle *BackwardsInvocation,
	request *dify_invocation.InvokeSpeech2TextRequest,
) {
	response, err := dify_invocation.InvokeSpeech2Text(request)
	if err != nil {
		handle.WriteError(fmt.Errorf("invoke speech2text model failed: %s", err.Error()))
		return
	}

	handle.WriteResponse("struct", response)
}

func executeDifyInvocationModerationTask(
	handle *BackwardsInvocation,
	request *dify_invocation.InvokeModerationRequest,
) {
	response, err := dify_invocation.InvokeModeration(request)
	if err != nil {
		handle.WriteError(fmt.Errorf("invoke moderation model failed: %s", err.Error()))
		return
	}

	handle.WriteResponse("struct", response)
}

package backwards_invocation

import (
	"encoding/hex"
	"fmt"

	"github.com/langgenius/dify-plugin-daemon/internal/core/dify_invocation"
	"github.com/langgenius/dify-plugin-daemon/internal/core/persistence"
	"github.com/langgenius/dify-plugin-daemon/internal/core/plugin_daemon/access_types"
	"github.com/langgenius/dify-plugin-daemon/internal/core/session_manager"
	"github.com/langgenius/dify-plugin-daemon/internal/types/entities/model_entities"
	"github.com/langgenius/dify-plugin-daemon/internal/types/entities/plugin_entities"
	"github.com/langgenius/dify-plugin-daemon/internal/types/entities/tool_entities"
	"github.com/langgenius/dify-plugin-daemon/internal/utils/parser"
	"github.com/langgenius/dify-plugin-daemon/internal/utils/routine"
)

// returns error only if payload is not correct
func InvokeDify(
	declaration *plugin_entities.PluginDeclaration,
	invoke_from access_types.PluginAccessType,
	session *session_manager.Session,
	writer BackwardsInvocationWriter,
	data []byte,
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
	request_handle, err := prepareDifyInvocationArguments(
		session,
		writer,
		request,
	)
	if err != nil {
		return err
	}

	if invoke_from == access_types.PLUGIN_ACCESS_TYPE_MODEL {
		request_handle.WriteError(fmt.Errorf("you can not invoke dify from %s", invoke_from))
		request_handle.EndResponse()
		return nil
	}

	// check permission
	if err := checkPermission(declaration, request_handle); err != nil {
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
			"func": func(declaration *plugin_entities.PluginDeclaration) bool {
				return declaration.Resource.Permission.AllowInvokeTool()
			},
			"error": "permission denied, you need to enable tool access in plugin manifest",
		},
		dify_invocation.INVOKE_TYPE_LLM: {
			"func": func(declaration *plugin_entities.PluginDeclaration) bool {
				return declaration.Resource.Permission.AllowInvokeLLM()
			},
			"error": "permission denied, you need to enable llm access in plugin manifest",
		},
		dify_invocation.INVOKE_TYPE_TEXT_EMBEDDING: {
			"func": func(declaration *plugin_entities.PluginDeclaration) bool {
				return declaration.Resource.Permission.AllowInvokeTextEmbedding()
			},
			"error": "permission denied, you need to enable text-embedding access in plugin manifest",
		},
		dify_invocation.INVOKE_TYPE_RERANK: {
			"func": func(declaration *plugin_entities.PluginDeclaration) bool {
				return declaration.Resource.Permission.AllowInvokeRerank()
			},
			"error": "permission denied, you need to enable rerank access in plugin manifest",
		},
		dify_invocation.INVOKE_TYPE_TTS: {
			"func": func(declaration *plugin_entities.PluginDeclaration) bool {
				return declaration.Resource.Permission.AllowInvokeTTS()
			},
			"error": "permission denied, you need to enable tts access in plugin manifest",
		},
		dify_invocation.INVOKE_TYPE_SPEECH2TEXT: {
			"func": func(declaration *plugin_entities.PluginDeclaration) bool {
				return declaration.Resource.Permission.AllowInvokeSpeech2Text()
			},
			"error": "permission denied, you need to enable speech2text access in plugin manifest",
		},
		dify_invocation.INVOKE_TYPE_MODERATION: {
			"func": func(declaration *plugin_entities.PluginDeclaration) bool {
				return declaration.Resource.Permission.AllowInvokeModeration()
			},
			"error": "permission denied, you need to enable moderation access in plugin manifest",
		},
		dify_invocation.INVOKE_TYPE_NODE: {
			"func": func(declaration *plugin_entities.PluginDeclaration) bool {
				return declaration.Resource.Permission.AllowInvokeNode()
			},
			"error": "permission denied, you need to enable node access in plugin manifest",
		},
		dify_invocation.INVOKE_TYPE_APP: {
			"func": func(declaration *plugin_entities.PluginDeclaration) bool {
				return declaration.Resource.Permission.AllowInvokeApp()
			},
			"error": "permission denied, you need to enable app access in plugin manifest",
		},
		dify_invocation.INVOKE_TYPE_STORAGE: {
			"func": func(declaration *plugin_entities.PluginDeclaration) bool {
				return declaration.Resource.Permission.AllowInvokeStorage()
			},
			"error": "permission denied, you need to enable storage access in plugin manifest",
		},
	}
)

func checkPermission(runtime *plugin_entities.PluginDeclaration, request_handle *BackwardsInvocation) error {
	permission, ok := permissionMapping[request_handle.Type()]
	if !ok {
		return fmt.Errorf("unsupported invoke type: %s", request_handle.Type())
	}

	permission_func, ok := permission["func"].(func(runtime *plugin_entities.PluginDeclaration) bool)
	if !ok {
		return fmt.Errorf("permission function not found: %s", request_handle.Type())
	}

	if !permission_func(runtime) {
		return fmt.Errorf(permission["error"].(string))
	}

	return nil
}

func prepareDifyInvocationArguments(
	session *session_manager.Session,
	writer BackwardsInvocationWriter,
	request map[string]any,
) (*BackwardsInvocation, error) {
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
		backwards_request_id,
		session,
		writer,
		detailed_request,
	), nil
}

var (
	dispatchMapping = map[dify_invocation.InvokeType]func(handle *BackwardsInvocation){
		dify_invocation.INVOKE_TYPE_TOOL: func(handle *BackwardsInvocation) {
			genericDispatchTask(handle, executeDifyInvocationToolTask)
		},
		dify_invocation.INVOKE_TYPE_LLM: func(handle *BackwardsInvocation) {
			genericDispatchTask(handle, executeDifyInvocationLLMTask)
		},
		dify_invocation.INVOKE_TYPE_TEXT_EMBEDDING: func(handle *BackwardsInvocation) {
			genericDispatchTask(handle, executeDifyInvocationTextEmbeddingTask)
		},
		dify_invocation.INVOKE_TYPE_RERANK: func(handle *BackwardsInvocation) {
			genericDispatchTask(handle, executeDifyInvocationRerankTask)
		},
		dify_invocation.INVOKE_TYPE_TTS: func(handle *BackwardsInvocation) {
			genericDispatchTask(handle, executeDifyInvocationTTSTask)
		},
		dify_invocation.INVOKE_TYPE_SPEECH2TEXT: func(handle *BackwardsInvocation) {
			genericDispatchTask(handle, executeDifyInvocationSpeech2TextTask)
		},
		dify_invocation.INVOKE_TYPE_MODERATION: func(handle *BackwardsInvocation) {
			genericDispatchTask(handle, executeDifyInvocationModerationTask)
		},
		dify_invocation.INVOKE_TYPE_APP: func(handle *BackwardsInvocation) {
			genericDispatchTask(handle, executeDifyInvocationAppTask)
		},
		dify_invocation.INVOKE_TYPE_STORAGE: func(handle *BackwardsInvocation) {
			genericDispatchTask(handle, executeDifyInvocationStorageTask)
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
	response, err := handle.backwardsInvocation.InvokeTool(request)
	if err != nil {
		handle.WriteError(fmt.Errorf("invoke tool failed: %s", err.Error()))
		return
	}

	response.Async(func(t tool_entities.ToolResponseChunk) {
		handle.WriteResponse("stream", t)
	})
}

func executeDifyInvocationLLMTask(
	handle *BackwardsInvocation,
	request *dify_invocation.InvokeLLMRequest,
) {
	response, err := handle.backwardsInvocation.InvokeLLM(request)
	if err != nil {
		handle.WriteError(fmt.Errorf("invoke llm model failed: %s", err.Error()))
		return
	}

	response.Async(func(t model_entities.LLMResultChunk) {
		handle.WriteResponse("stream", t)
	})
}

func executeDifyInvocationTextEmbeddingTask(
	handle *BackwardsInvocation,
	request *dify_invocation.InvokeTextEmbeddingRequest,
) {
	response, err := handle.backwardsInvocation.InvokeTextEmbedding(request)
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
	response, err := handle.backwardsInvocation.InvokeRerank(request)
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
	response, err := handle.backwardsInvocation.InvokeTTS(request)
	if err != nil {
		handle.WriteError(fmt.Errorf("invoke tts model failed: %s", err.Error()))
		return
	}

	response.Async(func(t model_entities.TTSResult) {
		handle.WriteResponse("struct", t)
	})
}

func executeDifyInvocationSpeech2TextTask(
	handle *BackwardsInvocation,
	request *dify_invocation.InvokeSpeech2TextRequest,
) {
	response, err := handle.backwardsInvocation.InvokeSpeech2Text(request)
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
	response, err := handle.backwardsInvocation.InvokeModeration(request)
	if err != nil {
		handle.WriteError(fmt.Errorf("invoke moderation model failed: %s", err.Error()))
		return
	}

	handle.WriteResponse("struct", response)
}

func executeDifyInvocationAppTask(
	handle *BackwardsInvocation,
	request *dify_invocation.InvokeAppRequest,
) {
	response, err := handle.backwardsInvocation.InvokeApp(request)
	if err != nil {
		handle.WriteError(fmt.Errorf("invoke app failed: %s", err.Error()))
		return
	}

	user_id, err := handle.UserID()
	if err != nil {
		handle.WriteError(fmt.Errorf("get user id failed: %s", err.Error()))
		return
	}

	request.User = user_id

	response.Async(func(t map[string]any) {
		handle.WriteResponse("stream", t)
	})
}

func executeDifyInvocationStorageTask(
	handle *BackwardsInvocation,
	request *dify_invocation.InvokeStorageRequest,
) {
	if handle.session == nil {
		handle.WriteError(fmt.Errorf("session not found"))
		return
	}

	persistence := persistence.GetPersistence()
	if persistence == nil {
		handle.WriteError(fmt.Errorf("persistence not found"))
		return
	}

	tenant_id, err := handle.TenantID()
	if err != nil {
		handle.WriteError(fmt.Errorf("get tenant id failed: %s", err.Error()))
		return
	}

	plugin_id := handle.session.PluginUniqueIdentifier

	if request.Opt == dify_invocation.STORAGE_OPT_GET {
		data, err := persistence.Load(tenant_id, plugin_id.PluginID(), request.Key)
		if err != nil {
			handle.WriteError(fmt.Errorf("load data failed: %s", err.Error()))
			return
		}

		handle.WriteResponse("struct", map[string]any{
			"data": hex.EncodeToString(data),
		})
	} else if request.Opt == dify_invocation.STORAGE_OPT_SET {
		data, err := hex.DecodeString(request.Value)
		if err != nil {
			handle.WriteError(fmt.Errorf("decode data failed: %s", err.Error()))
			return
		}

		if err := persistence.Save(tenant_id, plugin_id.PluginID(), request.Key, data); err != nil {
			handle.WriteError(fmt.Errorf("save data failed: %s", err.Error()))
			return
		}

		handle.WriteResponse("struct", map[string]any{
			"data": "ok",
		})
	} else if request.Opt == dify_invocation.STORAGE_OPT_DEL {
		if err := persistence.Delete(tenant_id, plugin_id.PluginID(), request.Key); err != nil {
			handle.WriteError(fmt.Errorf("delete data failed: %s", err.Error()))
			return
		}

		handle.WriteResponse("struct", map[string]any{
			"data": "ok",
		})
	}
}

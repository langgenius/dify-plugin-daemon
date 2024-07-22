package plugin_daemon

import (
	"fmt"

	"github.com/langgenius/dify-plugin-daemon/internal/core/dify_invocation"
	"github.com/langgenius/dify-plugin-daemon/internal/core/plugin_daemon/backwards_invocation"
	"github.com/langgenius/dify-plugin-daemon/internal/core/session_manager"
	"github.com/langgenius/dify-plugin-daemon/internal/types/entities"
	"github.com/langgenius/dify-plugin-daemon/internal/utils/log"
	"github.com/langgenius/dify-plugin-daemon/internal/utils/parser"
	"github.com/langgenius/dify-plugin-daemon/internal/utils/routine"
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
		r, err := parser.MapToStruct[dify_invocation.InvokeToolRequest](handle.RequestData())
		if err != nil {
			handle.WriteError(fmt.Errorf("unmarshal invoke tool request failed: %s", err.Error()))
			return
		}

		submitToolTask(runtime, session, backwards_request_id, &r)
	case dify_invocation.INVOKE_TYPE_MODEL:
		r, err := parser.MapToStruct[dify_invocation.InvokeModelRequest](handle.RequestData())
		if err != nil {
			handle.WriteError(fmt.Errorf("unmarshal invoke model request failed: %s", err.Error()))
			return
		}

		submitModelTask(runtime, session, backwards_request_id, &r)
	case dify_invocation.INVOKE_TYPE_NODE:
		node_type, ok := detailed_request["node_type"].(dify_invocation.NodeType)
		if !ok {
			return fmt.Errorf("invoke request missing node_type: %s", data)
		}
		node_data, ok := detailed_request["data"].(map[string]any)
		if !ok {
			return fmt.Errorf("invoke request missing data: %s", data)
		}
		switch node_type {
		case dify_invocation.QUESTION_CLASSIFIER:
			d := dify_invocation.InvokeNodeRequest[dify_invocation.QuestionClassifierNodeData]{
				NodeType: dify_invocation.QUESTION_CLASSIFIER,
			}
			if err := d.FromMap(node_data); err != nil {
				return fmt.Errorf("unmarshal question classifier node data failed: %s", err.Error())
			}
			submitNodeInvocationRequestTask(runtime, session, backwards_request_id, &d)
		case dify_invocation.KNOWLEDGE_RETRIEVAL:
			d := dify_invocation.InvokeNodeRequest[dify_invocation.KnowledgeRetrievalNodeData]{
				NodeType: dify_invocation.KNOWLEDGE_RETRIEVAL,
			}
			if err := d.FromMap(node_data); err != nil {
				return fmt.Errorf("unmarshal knowledge retrieval node data failed: %s", err.Error())
			}
			submitNodeInvocationRequestTask(runtime, session, backwards_request_id, &d)
		case dify_invocation.PARAMETER_EXTRACTOR:
			d := dify_invocation.InvokeNodeRequest[dify_invocation.ParameterExtractorNodeData]{
				NodeType: dify_invocation.PARAMETER_EXTRACTOR,
			}
			if err := d.FromMap(node_data); err != nil {
				return fmt.Errorf("unmarshal parameter extractor node data failed: %s", err.Error())
			}
			submitNodeInvocationRequestTask(runtime, session, backwards_request_id, &d)
		default:
			return fmt.Errorf("unknown node type: %s", node_type)
		}
	default:
		return fmt.Errorf("unknown invoke type: %s", typ)
	}
}

func setTaskContext(session *session_manager.Session, r *dify_invocation.BaseInvokeDifyRequest) {
	r.TenantId = session.TenantID()
	r.UserId = session.UserID()
}

func submitModelTask(
	runtime entities.PluginRuntimeInterface,
	session *session_manager.Session,
	request_id string,
	t *dify_invocation.InvokeModelRequest,
) {
	setTaskContext(session, &t.BaseInvokeDifyRequest)
	routine.Submit(func() {
		response, err := dify_invocation.InvokeModel(t)
		if err != nil {
			log.Error("invoke model failed: %s", err.Error())
			return
		}
		defer response.Close()

		for response.Next() {
			chunk, _ := response.Read()
			fmt.Println(chunk)
		}
	})
}

func submitToolTask(
	runtime entities.PluginRuntimeInterface,
	session *session_manager.Session,
	request_id string,
	t *dify_invocation.InvokeToolRequest,
) {
	setTaskContext(session, &t.BaseInvokeDifyRequest)
	routine.Submit(func() {
		response, err := dify_invocation.InvokeTool(t)
		if err != nil {
			log.Error("invoke tool failed: %s", err.Error())
			return
		}
		defer response.Close()

		for response.Next() {
			chunk, _ := response.Read()
			fmt.Println(chunk)
		}
	})
}

func submitNodeInvocationRequestTask[W dify_invocation.WorkflowNodeData](
	runtime entities.PluginRuntimeInterface,
	session *session_manager.Session,
	request_id string,
	t *dify_invocation.InvokeNodeRequest[W],
) {
	setTaskContext(session, &t.BaseInvokeDifyRequest)
	routine.Submit(func() {
		response, err := dify_invocation.InvokeNode(t)
		if err != nil {
			log.Error("invoke node failed: %s", err.Error())
			return
		}

		fmt.Println(response)
	})
}

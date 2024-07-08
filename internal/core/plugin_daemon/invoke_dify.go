package plugin_daemon

import (
	"fmt"

	"github.com/langgenius/dify-plugin-daemon/internal/core/dify_invocation"
	"github.com/langgenius/dify-plugin-daemon/internal/core/session_manager"
	"github.com/langgenius/dify-plugin-daemon/internal/types/entities"
	"github.com/langgenius/dify-plugin-daemon/internal/utils/log"
	"github.com/langgenius/dify-plugin-daemon/internal/utils/parser"
	"github.com/langgenius/dify-plugin-daemon/internal/utils/routine"
)

func invokeDify(runtime entities.PluginRuntimeInterface,
	session *session_manager.Session, data []byte,
) error {
	// unmarshal invoke data
	request, err := parser.UnmarshalJsonBytes[map[string]any](data)

	if err != nil {
		return fmt.Errorf("unmarshal invoke request failed: %s", err.Error())
	}

	typ, ok := request["type"].(string)
	if !ok {
		return fmt.Errorf("invoke request missing type: %s", data)
	}

	// get request id
	request_id, ok := request["request_id"].(string)
	if !ok {
		return fmt.Errorf("invoke request missing request_id: %s", data)
	}

	// get request
	detailed_request, ok := request["request"].(map[string]any)
	if !ok {
		return fmt.Errorf("invoke request missing request: %s", data)
	}

	switch typ {
	case "tool":
		r := dify_invocation.InvokeToolRequest{}
		if err := r.FromMap(request, detailed_request); err != nil {
			return fmt.Errorf("unmarshal tool invoke request failed: %s", err.Error())
		}
		submitToolTask(runtime, session, request_id, &r)
	case "model":
		r := dify_invocation.InvokeModelRequest{}
		if err := r.FromMap(request, detailed_request); err != nil {
			return fmt.Errorf("unmarshal model invoke request failed: %s", err.Error())
		}
		submitModelTask(runtime, session, request_id, &r)
	case "node":
		node_type, ok := detailed_request["node_type"].(string)
		if !ok {
			return fmt.Errorf("invoke request missing node_type: %s", data)
		}
		node_data, ok := detailed_request["data"].(map[string]any)
		if !ok {
			return fmt.Errorf("invoke request missing data: %s", data)
		}
		switch node_type {
		case dify_invocation.NODE_TYPE_QUESTION_CLASSIFIER:
			d := dify_invocation.InvokeNodeRequest[*dify_invocation.QuestionClassifierNodeData]{
				NodeType: dify_invocation.NODE_TYPE_QUESTION_CLASSIFIER,
				NodeData: &dify_invocation.QuestionClassifierNodeData{},
			}
			if err := d.FromMap(node_data); err != nil {
				return fmt.Errorf("unmarshal question classifier node data failed: %s", err.Error())
			}
			submitNodeInvocationRequestTask(runtime, session, request_id, &d)
		case dify_invocation.NODE_TYPE_KNOWLEDGE_RETRIEVAL:
			d := dify_invocation.InvokeNodeRequest[*dify_invocation.KnowledgeRetrievalNodeData]{
				NodeType: dify_invocation.NODE_TYPE_KNOWLEDGE_RETRIEVAL,
				NodeData: &dify_invocation.KnowledgeRetrievalNodeData{},
			}
			if err := d.FromMap(node_data); err != nil {
				return fmt.Errorf("unmarshal knowledge retrieval node data failed: %s", err.Error())
			}
			submitNodeInvocationRequestTask(runtime, session, request_id, &d)
		case dify_invocation.NODE_TYPE_PARAMETER_EXTRACTOR:
			d := dify_invocation.InvokeNodeRequest[*dify_invocation.ParameterExtractorNodeData]{
				NodeType: dify_invocation.NODE_TYPE_PARAMETER_EXTRACTOR,
				NodeData: &dify_invocation.ParameterExtractorNodeData{},
			}
			if err := d.FromMap(node_data); err != nil {
				return fmt.Errorf("unmarshal parameter extractor node data failed: %s", err.Error())
			}
			submitNodeInvocationRequestTask(runtime, session, request_id, &d)
		case dify_invocation.NODE_TYPE_CODE:
			d := dify_invocation.InvokeNodeRequest[*dify_invocation.CodeNodeData]{
				NodeType: dify_invocation.NODE_TYPE_CODE,
				NodeData: &dify_invocation.CodeNodeData{},
			}
			if err := d.FromMap(node_data); err != nil {
				return fmt.Errorf("unmarshal code node data failed: %s", err.Error())
			}
			submitNodeInvocationRequestTask(runtime, session, request_id, &d)
		default:
			return fmt.Errorf("unknown node type: %s", node_type)
		}
	default:
		return fmt.Errorf("unknown invoke type: %s", typ)
	}

	return nil
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

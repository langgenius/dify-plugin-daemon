package dify_invocation

import (
	"encoding/json"

	"github.com/langgenius/dify-plugin-daemon/internal/types/entities/model_entities"
)

type BaseInvokeDifyRequest struct {
	TenantId string     `json:"tenant_id"`
	UserId   string     `json:"user_id"`
	Type     InvokeType `json:"type"`
}

type InvokeType string

const (
	INVOKE_TYPE_MODEL InvokeType = "model"
	INVOKE_TYPE_TOOL  InvokeType = "tool"
	INVOKE_TYPE_NODE  InvokeType = "node"
)

type InvokeModelRequest struct {
	BaseInvokeDifyRequest
	Provider   string                   `json:"provider"`
	Model      string                   `json:"model"`
	ModelType  model_entities.ModelType `json:"model_type"`
	Parameters map[string]any           `json:"parameters"`
}

func (r InvokeModelRequest) MarshalJSON() ([]byte, error) {
	flattened := make(map[string]any)
	flattened["tenant_id"] = r.TenantId
	flattened["user_id"] = r.UserId
	flattened["provider"] = r.Provider
	flattened["model"] = r.Model
	flattened["parameters"] = r.Parameters
	return json.Marshal(flattened)
}

type InvokeModelResponseChunk struct {
}

type InvokeToolRequest struct {
	BaseInvokeDifyRequest
	Provider   string         `json:"provider"`
	Tool       string         `json:"tool"`
	Parameters map[string]any `json:"parameters"`
}

func (r InvokeToolRequest) MarshalJSON() ([]byte, error) {
	flattened := make(map[string]any)
	flattened["tenant_id"] = r.TenantId
	flattened["user_id"] = r.UserId
	flattened["provider"] = r.Provider
	flattened["tool"] = r.Tool
	flattened["parameters"] = r.Parameters
	return json.Marshal(flattened)
}

type InvokeToolResponseChunk struct {
}

type InvokeNodeRequest[T WorkflowNodeData] struct {
	BaseInvokeDifyRequest
	NodeType NodeType `json:"node_type"`
	NodeData T        `json:"node_data"`
}

func (r InvokeNodeRequest[T]) MarshalJSON() ([]byte, error) {
	flattened := make(map[string]any)
	flattened["tenant_id"] = r.TenantId
	flattened["user_id"] = r.UserId
	flattened["node_type"] = r.NodeType
	flattened["node_data"] = r.NodeData
	return json.Marshal(flattened)
}

type InvokeNodeResponse struct {
}

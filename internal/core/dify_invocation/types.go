package dify_invocation

import "encoding/json"

type BaseInvokeDifyRequest struct {
	TenantId string `json:"tenant_id"`
	UserId   string `json:"user_id"`
}

type InvokeModelRequest struct {
	BaseInvokeDifyRequest
	Provider   string         `json:"provider"`
	Model      string         `json:"model"`
	Parameters map[string]any `json:"parameters"`
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
	NodeType string `json:"node_type"`
	NodeData T      `json:"node_data"`
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

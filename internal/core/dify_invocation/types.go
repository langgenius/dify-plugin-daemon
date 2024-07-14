package dify_invocation

import (
	"encoding/json"
	"fmt"

	"github.com/langgenius/dify-plugin-daemon/internal/types/entities/model_entities"
)

type BaseInvokeDifyRequest struct {
	TenantId string     `json:"tenant_id"`
	UserId   string     `json:"user_id"`
	Type     InvokeType `json:"type"`
}

func (r *BaseInvokeDifyRequest) FromMap(data map[string]any) error {
	var ok bool
	if r.TenantId, ok = data["tenant_id"].(string); !ok {
		return fmt.Errorf("tenant_id is not a string")
	}

	if r.UserId, ok = data["user_id"].(string); !ok {
		return fmt.Errorf("user_id is not a string")
	}

	if r.Type, ok = data["type"].(InvokeType); !ok {
		return fmt.Errorf("type is not a string")
	}

	return nil
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

func (r *InvokeModelRequest) FromMap(base map[string]any, data map[string]any) error {
	var ok bool
	if r.Provider, ok = data["provider"].(string); !ok {
		return fmt.Errorf("provider is not a string")
	}

	if r.Model, ok = data["model"].(string); !ok {
		return fmt.Errorf("model is not a string")
	}

	if r.ModelType, ok = data["model_type"].(model_entities.ModelType); !ok {
		return fmt.Errorf("model_type is not a string")
	}

	if r.Parameters, ok = data["parameters"].(map[string]any); !ok {
		return fmt.Errorf("parameters is not a map")
	}

	return nil
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

func (r *InvokeToolRequest) FromMap(base map[string]any, data map[string]any) error {
	var ok bool
	if r.Provider, ok = data["provider"].(string); !ok {
		return fmt.Errorf("provider is not a string")
	}

	if r.Tool, ok = data["tool"].(string); !ok {
		return fmt.Errorf("tool is not a string")
	}

	if r.Parameters, ok = data["parameters"].(map[string]any); !ok {
		return fmt.Errorf("parameters is not a map")
	}

	return nil
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

func (r *InvokeNodeRequest[T]) FromMap(data map[string]any) error {
	var ok bool
	if r.NodeType, ok = data["node_type"].(NodeType); !ok {
		return fmt.Errorf("node_type is not a string")
	}

	if err := r.NodeData.FromMap(data["node_data"].(map[string]any)); err != nil {
		return err
	}

	return nil
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

package agent_entities

import "github.com/langgenius/dify-plugin-daemon/pkg/entities/tool_entities"

type AgentStrategyResponseChunk struct {
	tool_entities.ToolResponseChunk `json:",inline"`
}

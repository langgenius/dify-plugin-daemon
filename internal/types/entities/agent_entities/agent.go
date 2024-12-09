package agent_entities

import "github.com/langgenius/dify-plugin-daemon/internal/types/entities/tool_entities"

type AgentResponseChunk struct {
	tool_entities.ToolResponseChunk `json:",inline"`
}

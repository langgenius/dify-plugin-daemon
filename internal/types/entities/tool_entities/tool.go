package tool_entities

type ToolResponseChunk struct {
	Type    string         `json:"type"`
	Message map[string]any `json:"message"`
}

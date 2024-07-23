package requests

type InvokeToolSchema struct {
	Provider       string         `json:"provider" validate:"required"`
	Tool           string         `json:"tool" validate:"required"`
	ToolParameters map[string]any `json:"tool_parameters" validate:"omitempty,dive,is_basic_type"`
}

type RequestInvokeTool struct {
	InvokeToolSchema
	Credentials
}

type RequestValidateToolCredentials struct {
	Provider    string         `json:"provider" validate:"required"`
	Credentials map[string]any `json:"credentials" validate:"omitempty,dive,is_basic_type"`
}

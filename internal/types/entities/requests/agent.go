package requests

type InvokeAgentSchema struct {
	Provider    string         `json:"provider" validate:"required"`
	Strategy    string         `json:"strategy" validate:"required"`
	AgentParams map[string]any `json:"agent_params" validate:"omitempty"`
}

type RequestInvokeAgent struct {
	InvokeAgentSchema
}

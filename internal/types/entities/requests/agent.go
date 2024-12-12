package requests

type InvokeAgentStrategySchema struct {
	AgentStrategyProvider string         `json:"agent_strategy_provider" validate:"required"`
	AgentStrategy         string         `json:"agent_strategy" validate:"required"`
	AgentStrategyParams   map[string]any `json:"agent_strategy_params" validate:"omitempty"`
}

type RequestInvokeAgentStrategy struct {
	InvokeAgentStrategySchema
}

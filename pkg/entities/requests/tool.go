package requests

import (
	"github.com/go-playground/validator/v10"
	"github.com/langgenius/dify-plugin-daemon/pkg/validators"
)

type ToolType string

const (
	TOOL_TYPE_BUILTIN  ToolType = "builtin"
	TOOL_TYPE_WORKFLOW ToolType = "workflow"
	TOOL_TYPE_API      ToolType = "api"
	TOOL_TYPE_MCP      ToolType = "mcp"
)

func init() {
	validators.GlobalEntitiesValidator.RegisterValidation("tool_type", func(fl validator.FieldLevel) bool {
		switch fl.Field().String() {
		case string(TOOL_TYPE_BUILTIN), string(TOOL_TYPE_WORKFLOW), string(TOOL_TYPE_API), string(TOOL_TYPE_MCP):
			return true
		}
		return false
	})
}

type InvokeToolSchema struct {
	Provider       string         `json:"provider" validate:"required"`
	Tool           string         `json:"tool" validate:"required"`
	ToolParameters map[string]any `json:"tool_parameters" validate:"omitempty"`
}

type RequestInvokeTool struct {
	InvokeToolSchema
	Credentials
}

type RequestValidateToolCredentials struct {
	Provider    string         `json:"provider" validate:"required"`
	Credentials map[string]any `json:"credentials" validate:"omitempty"`
}

type RequestGetToolRuntimeParameters struct {
	Provider    string         `json:"provider" validate:"required"`
	Tool        string         `json:"tool" validate:"required"`
	Credentials map[string]any `json:"credentials" validate:"omitempty"`
}

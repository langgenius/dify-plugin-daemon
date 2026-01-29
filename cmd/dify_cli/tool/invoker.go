package tool

import (
	"fmt"

	"github.com/langgenius/dify-plugin-daemon/cmd/dify_cli/config"
	"github.com/langgenius/dify-plugin-daemon/cmd/dify_cli/types"
)

type Invoker interface {
	ShowHelp()
	PrepareParams(args []string) (map[string]any, error)
	GetCredentialID() string
	GetTool() *types.DifyToolDeclaration
}

func NewInvoker(cfg *types.DifyConfig, name string) (Invoker, error) {
	ref := config.FindToolReference(cfg, name)
	if ref == nil {
		return nil, fmt.Errorf("tool reference not found: %s (must use format: tool_name_uuid)", name)
	}

	tool := config.FindToolByReference(cfg, ref)
	if tool == nil {
		tools, err := FetchToolsBatch(cfg, []types.ToolReference{*ref})
		if err != nil {
			return nil, fmt.Errorf("failed to fetch tool info: %w", err)
		}
		if len(tools) == 0 {
			return nil, fmt.Errorf("tool '%s' from provider '%s' not found on server", ref.ToolName, ref.ToolProvider)
		}

		cfg.Tools = append(cfg.Tools, tools[0])
		if err := config.Save(cfg); err != nil {
			return nil, fmt.Errorf("failed to save config: %w", err)
		}
		tool = &cfg.Tools[len(cfg.Tools)-1]
	}

	if tool.Enabled != nil && !*tool.Enabled {
		return nil, fmt.Errorf("tool '%s' has been disabled by the user, you are not allowed to use it", ref.ToolName)
	}

	return &ReferenceInvoker{tool: tool, ref: ref}, nil
}

type ReferenceInvoker struct {
	tool *types.DifyToolDeclaration
	ref  *types.ToolReference
}

func (r *ReferenceInvoker) ShowHelp() {
	PrintHelp(r.tool, r.ref.DefaultValue)
}

func (r *ReferenceInvoker) PrepareParams(args []string) (map[string]any, error) {
	params := ParseArgs(r.tool, args)

	for paramName := range params {
		if fixedValue, isFixed := r.ref.DefaultValue[paramName]; isFixed {
			return nil, fmt.Errorf("parameter '%s' is fixed to '%v' and cannot be modified", paramName, fixedValue)
		}
	}

	for k, v := range r.ref.DefaultValue {
		params[k] = v
	}

	return params, nil
}

func (r *ReferenceInvoker) GetCredentialID() string {
	if r.ref.CredentialID != nil && *r.ref.CredentialID != "" {
		return *r.ref.CredentialID
	}
	return r.tool.CredentialId
}

func (r *ReferenceInvoker) GetTool() *types.DifyToolDeclaration {
	return r.tool
}

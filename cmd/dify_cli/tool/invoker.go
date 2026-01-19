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
	if ref := config.FindToolReference(cfg, name); ref != nil {
		tool := config.FindToolByReference(cfg, ref)
		if tool == nil {
			return nil, fmt.Errorf("referenced tool '%s' from provider '%s' not found", ref.ToolName, ref.ToolProvider)
		}
		return &ReferenceInvoker{tool: tool, ref: ref}, nil
	}

	tool := config.FindTool(cfg, name)
	if tool == nil {
		return nil, fmt.Errorf("tool not found: %s", name)
	}
	return &DirectInvoker{tool: tool}, nil
}

// DirectInvoker handles direct tool invocation
type DirectInvoker struct {
	tool *types.DifyToolDeclaration
}

func (d *DirectInvoker) ShowHelp() {
	PrintHelp(d.tool, nil)
}

func (d *DirectInvoker) PrepareParams(args []string) (map[string]any, error) {
	return ParseArgs(d.tool, args), nil
}

func (d *DirectInvoker) GetCredentialID() string {
	return d.tool.CredentialId
}

func (d *DirectInvoker) GetTool() *types.DifyToolDeclaration {
	return d.tool
}

// ReferenceInvoker handles tool reference invocation with fixed params
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
	if r.ref.CredentialID != "" {
		return r.ref.CredentialID
	}
	return r.tool.CredentialId
}

func (r *ReferenceInvoker) GetTool() *types.DifyToolDeclaration {
	return r.tool
}

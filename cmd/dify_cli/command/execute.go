package command

import (
	"fmt"
	"net/http"
	"os"
	"strings"

	"github.com/langgenius/dify-plugin-daemon/cmd/dify_cli/config"
	"github.com/langgenius/dify-plugin-daemon/cmd/dify_cli/types"
	"github.com/langgenius/dify-plugin-daemon/internal/core/dify_invocation"
	"github.com/langgenius/dify-plugin-daemon/pkg/entities/plugin_entities"
	"github.com/langgenius/dify-plugin-daemon/pkg/entities/requests"
	"github.com/langgenius/dify-plugin-daemon/pkg/entities/tool_entities"
	"github.com/langgenius/dify-plugin-daemon/pkg/utils/http_requests"
	"github.com/spf13/cobra"
)

var ExecuteCmd = &cobra.Command{
	Use:                "execute <tool_name> [--param value ...]",
	Short:              "Execute a tool directly",
	Long:               `Execute a tool directly by name. Useful for debugging.`,
	Example:            `  dify execute google_search --query "hello world"`,
	Args:               cobra.MinimumNArgs(1),
	DisableFlagParsing: true,
	Run:                runExecute,
}

func runExecute(cmd *cobra.Command, args []string) {
	toolName := args[0]
	toolArgs := args[1:]
	InvokeTool(toolName, toolArgs)
}

func InvokeTool(name string, args []string) {
	cfg, err := config.Load()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: config not found, run 'dify init' first: %v\n", err)
		os.Exit(1)
	}

	tool := config.FindTool(cfg, name)
	if tool == nil {
		fmt.Fprintf(os.Stderr, "Error: tool not found: %s\n", name)
		os.Exit(1)
	}

	params := parseToolArgs(tool, args)

	err = callDifyAPI(cfg, tool, params)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: API call failed: %v\n", err)
		os.Exit(1)
	}
}

func parseToolArgs(tool *types.DifyToolDeclaration, args []string) map[string]any {
	params := make(map[string]any)

	for i := 0; i < len(args); i++ {
		arg := args[i]
		if !strings.HasPrefix(arg, "--") {
			continue
		}

		name := strings.TrimPrefix(arg, "--")
		if i+1 >= len(args) {
			continue
		}

		value := args[i+1]
		i++

		for _, p := range tool.Parameters {
			if p.Name == name {
				switch p.Type {
				case plugin_entities.TOOL_PARAMETER_TYPE_NUMBER:
					var num float64
					fmt.Sscanf(value, "%f", &num)
					params[name] = num
				case plugin_entities.TOOL_PARAMETER_TYPE_BOOLEAN:
					params[name] = value == "true" || value == "1"
				default:
					params[name] = value
				}
				break
			}
		}

		if _, exists := params[name]; !exists {
			params[name] = value
		}
	}

	return params
}

func callDifyAPI(cfg *types.DifyConfig, tool *types.DifyToolDeclaration, params map[string]any) error {
	reqBody := dify_invocation.InvokeToolRequest{
		BaseInvokeDifyRequest: dify_invocation.BaseInvokeDifyRequest{
			TenantId: cfg.Env.TenantID,
			UserId:   cfg.Env.UserID,
			Type:     dify_invocation.INVOKE_TYPE_TOOL,
		},
		ToolType: tool.ProviderType,
		InvokeToolSchema: requests.InvokeToolSchema{
			Provider:       tool.Identity.Provider,
			Tool:           tool.Identity.Name,
			ToolParameters: params,
		},
		CredentialId:   tool.CredentialId,
		CredentialType: tool.CredentialType,
	}

	url := strings.TrimSuffix(cfg.Env.InnerAPIURL, "/") + "/inner/api/invoke/tool"
	client := &http.Client{}

	response, err := http_requests.PostAndParseStream[tool_entities.ToolResponseChunk](
		client,
		url,
		http_requests.HttpHeader(map[string]string{
			"X-Inner-Api-Key": cfg.Env.InnerAPIKey,
		}),
		http_requests.HttpPayloadJson(reqBody),
		http_requests.HttpUsingLengthPrefixed(true),
	)
	if err != nil {
		return err
	}
	defer response.Close()

	for response.Next() {
		chunk, err := response.Read()
		if err != nil {
			return err
		}

		if chunk.Type == tool_entities.ToolResponseChunkTypeText {
			if msg, ok := chunk.Message["text"]; ok {
				fmt.Print(msg)
			}
		}
	}
	return nil
}

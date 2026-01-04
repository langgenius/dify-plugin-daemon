package command

import (
	"errors"
	"fmt"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/langgenius/dify-plugin-daemon/cmd/dify_cli/config"
	toolhandler "github.com/langgenius/dify-plugin-daemon/cmd/dify_cli/tool"
	"github.com/langgenius/dify-plugin-daemon/cmd/dify_cli/types"
	"github.com/langgenius/dify-plugin-daemon/pkg/utils/http_requests"
	"github.com/langgenius/dify-plugin-daemon/pkg/utils/parser"
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

	var tool = config.FindTool(cfg, name)
	if tool == nil {
		fmt.Fprintf(os.Stderr, "Error: tool not found: %s\n", name)
		os.Exit(1)
	}

	if len(args) == 0 || args[0] == "--help" || args[0] == "-h" {
		PrintToolHelp(tool)
		return
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
				case types.ToolParameterTypeNumber:
					var num float64
					fmt.Sscanf(value, "%f", &num)
					params[name] = num
				case types.ToolParameterTypeBoolean:
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
	if cfg.Env.FilesURL != "" {
		toolhandler.SetFilesURL(cfg.Env.FilesURL)
	}

	reqBody := types.InvokeToolRequest{
		Type:           types.INVOKE_TYPE_TOOL,
		ToolType:       tool.ProviderType,
		Provider:       tool.Identity.Provider,
		Tool:           tool.Identity.Name,
		ToolParameters: params,
		CredentialId:   tool.CredentialId,
		CredentialType: tool.CredentialType,
	}

	url := strings.TrimSuffix(cfg.Env.InnerAPIURL, "/") + "/inner/api/invoke/tool"

	client := &http.Client{
		Timeout: 5 * time.Minute,
	}

	resp, err := http_requests.Request(
		client,
		url,
		"POST",
		http_requests.HttpHeader(map[string]string{
			"X-Inner-Api-Session-Id": cfg.Env.InnerAPISessionID,
		}),
		http_requests.HttpPayloadJson(reqBody),
		http_requests.HttpWriteTimeout(5000),
		http_requests.HttpReadTimeout(240000),
	)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	err = parser.LengthPrefixedChunking(resp.Body, 0x0f, 1024*1024*30, func(data []byte) error {
		chunk, err := parser.UnmarshalJsonBytes[types.DifyInnerAPIResponse[types.DifyToolResponseChunk]](data)
		if err != nil {
			return fmt.Errorf("unmarshal json failed: %v", err)
		}

		if chunk.Error != "" {
			return fmt.Errorf("API error: %s", chunk.Error)
		}

		if chunk.Data == nil {
			return errors.New("data is nil")
		}

		return toolhandler.Dispatch(chunk.Data)
	})

	return err
}

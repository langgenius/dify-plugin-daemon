package command

import (
	"errors"
	"fmt"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/langgenius/dify-plugin-daemon/cmd/dify_cli/config"
	"github.com/langgenius/dify-plugin-daemon/cmd/dify_cli/types"
	"github.com/langgenius/dify-plugin-daemon/internal/core/dify_invocation"
	"github.com/langgenius/dify-plugin-daemon/pkg/entities/plugin_entities"
	"github.com/langgenius/dify-plugin-daemon/pkg/entities/requests"
	"github.com/langgenius/dify-plugin-daemon/pkg/entities/tool_entities"
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

	client := &http.Client{
		Timeout: 5 * time.Minute,
	}

	resp, err := http_requests.Request(
		client,
		url,
		"POST",
		http_requests.HttpHeader(map[string]string{
			"X-Inner-Api-Key": cfg.Env.InnerAPIKey,
		}),
		http_requests.HttpPayloadJson(reqBody),
		http_requests.HttpWriteTimeout(5000),
		http_requests.HttpReadTimeout(240000),
	)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	type apiResponse struct {
		Data  *tool_entities.ToolResponseChunk `json:"data,omitempty"`
		Error string                           `json:"error"`
	}

	err = parser.LengthPrefixedChunking(resp.Body, 0x0f, 1024*1024*30, func(data []byte) error {
		chunk, err := parser.UnmarshalJsonBytes[apiResponse](data)
		if err != nil {
			return err
		}

		if chunk.Error != "" {
			return fmt.Errorf("API error: %s", chunk.Error)
		}

		if chunk.Data == nil {
			return errors.New("data is nil")
		}

		if chunk.Data.Type == tool_entities.ToolResponseChunkTypeText {
			if msg, ok := chunk.Data.Message["text"]; ok {
				fmt.Println(msg)
			}
		}
		if chunk.Data.Type == tool_entities.ToolResponseChunkTypeJson {
			if msg, ok := chunk.Data.Message["json"]; ok {
				fmt.Println(msg)
			}
		}
		if chunk.Data.Type == tool_entities.ToolResponseChunkTypeFile {
			if msg, ok := chunk.Data.Message["file"]; ok {
				fmt.Println(msg)
			}
		}
		if chunk.Data.Type == tool_entities.ToolResponseChunkTypeBlob {
			if msg, ok := chunk.Data.Message["blob"]; ok {
				fmt.Println(msg)
			}
		}
		if chunk.Data.Type == tool_entities.ToolResponseChunkTypeBlobChunk {
			if msg, ok := chunk.Data.Message["blob_chunk"]; ok {
				fmt.Println(msg)
			}
		}
		if chunk.Data.Type == tool_entities.ToolResponseChunkTypeLink {
			if msg, ok := chunk.Data.Message["link"]; ok {
				fmt.Println(msg)
			}
		}
		return nil
	})

	fmt.Println()
	return err
}

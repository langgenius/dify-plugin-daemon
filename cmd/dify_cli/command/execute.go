package command

import (
	"bytes"
	"encoding/binary"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"

	"github.com/langgenius/dify-plugin-daemon/cmd/dify_cli/config"
	"github.com/langgenius/dify-plugin-daemon/cmd/dify_cli/types"
	"github.com/langgenius/dify-plugin-daemon/pkg/entities/plugin_entities"
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

	provider, tool := config.FindTool(cfg, name)
	if tool == nil {
		fmt.Fprintf(os.Stderr, "Error: tool not found: %s\n", name)
		os.Exit(1)
	}

	params := parseToolArgs(tool, args)

	response, err := callDifyAPI(cfg, provider.Identity.Name, tool.Identity.Name, params)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: API call failed: %v\n", err)
		os.Exit(1)
	}

	fmt.Print(response)
}

func parseToolArgs(tool *plugin_entities.ToolDeclaration, args []string) map[string]any {
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

func callDifyAPI(cfg *types.DifyConfig, providerName, toolName string, params map[string]any) (string, error) {
	reqBody := map[string]any{
		"provider":        providerName,
		"tool":            toolName,
		"tool_parameters": params,
	}

	jsonData, err := json.Marshal(reqBody)
	if err != nil {
		return "", err
	}

	url := strings.TrimSuffix(cfg.Env.InnerAPIURL, "/") + "/inner/api/invoke/tool"
	req, err := http.NewRequest("POST", url, bytes.NewReader(jsonData))
	if err != nil {
		return "", err
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Inner-Api-Key", cfg.Env.InnerAPIKey)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("HTTP %d: %s", resp.StatusCode, string(body))
	}

	return parseStreamResponse(resp.Body)
}

func parseStreamResponse(reader io.Reader) (string, error) {
	var result strings.Builder

	for {
		var marker byte
		if err := binary.Read(reader, binary.BigEndian, &marker); err != nil {
			if err == io.EOF {
				break
			}
			return result.String(), err
		}

		if marker != 0x0f {
			continue
		}

		var length uint32
		if err := binary.Read(reader, binary.BigEndian, &length); err != nil {
			return result.String(), err
		}

		data := make([]byte, length)
		if _, err := io.ReadFull(reader, data); err != nil {
			return result.String(), err
		}

		var chunk struct {
			Type    string `json:"type"`
			Message string `json:"message"`
		}

		if err := json.Unmarshal(data, &chunk); err != nil {
			continue
		}

		if chunk.Type == "text" {
			result.WriteString(chunk.Message)
		}
	}

	return result.String(), nil
}

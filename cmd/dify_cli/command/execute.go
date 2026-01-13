package command

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/langgenius/dify-plugin-daemon/cmd/dify_cli/config"
	toolhandler "github.com/langgenius/dify-plugin-daemon/cmd/dify_cli/tool"
	"github.com/langgenius/dify-plugin-daemon/cmd/dify_cli/types"
	"github.com/langgenius/dify-plugin-daemon/cmd/dify_cli/uploader"
	"github.com/langgenius/dify-plugin-daemon/pkg/utils/encryption"
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

	params, fileParams := parseToolArgs(tool, args)

	err = callDifyAPI(cfg, tool, params, fileParams)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: API call failed: %v\n", err)
		os.Exit(1)
	}
}

type fileParamInfo struct {
	paramName string
	paramType types.ToolParameterType
	paths     []string
}

func parseToolArgs(tool *types.DifyToolDeclaration, args []string) (map[string]any, []fileParamInfo) {
	params := make(map[string]any)
	var fileParams []fileParamInfo

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
				case types.ToolParameterTypeFile:
					fileParams = append(fileParams, fileParamInfo{
						paramName: name,
						paramType: p.Type,
						paths:     []string{value},
					})
				case types.ToolParameterTypeFiles:
					paths := strings.Split(value, ",")
					for j := range paths {
						paths[j] = strings.TrimSpace(paths[j])
					}
					fileParams = append(fileParams, fileParamInfo{
						paramName: name,
						paramType: p.Type,
						paths:     paths,
					})
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

	return params, fileParams
}

func signRequest(secret string, timestamp string, body []byte) string {
	data := append([]byte(timestamp+"."), body...)
	return "sha256=" + encryption.HmacSha256(secret, data)
}

func callDifyAPI(cfg *types.DifyConfig, tool *types.DifyToolDeclaration, params map[string]any, fileParams []fileParamInfo) error {
	if cfg.Env.FilesURL != "" {
		toolhandler.SetFilesURL(cfg.Env.FilesURL)
	}

	for _, fp := range fileParams {
		if fp.paramType == types.ToolParameterTypeFile {
			if len(fp.paths) == 0 {
				continue
			}
			fileObj, err := uploader.UploadFile(cfg, fp.paths[0])
			if err != nil {
				return fmt.Errorf("failed to upload file for parameter '%s': %w", fp.paramName, err)
			}
			params[fp.paramName] = fileObj
		} else if fp.paramType == types.ToolParameterTypeFiles {
			var fileObjs []*types.ToolFileObject
			for _, path := range fp.paths {
				fileObj, err := uploader.UploadFile(cfg, path)
				if err != nil {
					return fmt.Errorf("failed to upload file '%s' for parameter '%s': %w", path, fp.paramName, err)
				}
				fileObjs = append(fileObjs, fileObj)
			}
			params[fp.paramName] = fileObjs
		}
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

	body, err := json.Marshal(reqBody)
	if err != nil {
		return err
	}

	timestamp := strconv.FormatInt(time.Now().Unix(), 10)
	signature := signRequest(cfg.Env.CliApiSecret, timestamp, body)

	url := strings.TrimSuffix(cfg.Env.CliApiURL, "/") + "/cli/api/invoke/tool"

	client := &http.Client{
		Timeout: 5 * time.Minute,
	}

	resp, err := http_requests.Request(
		client,
		url,
		"POST",
		http_requests.HttpHeader(map[string]string{
			"X-Cli-Api-Session-Id": cfg.Env.CliApiSessionID,
			"X-Cli-Api-Timestamp":  timestamp,
			"X-Cli-Api-Signature":  signature,
		}),
		http_requests.HttpPayloadJson(reqBody),
		http_requests.HttpWriteTimeout(5000),
		http_requests.HttpReadTimeout(240000),
	)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		bodyBytes, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("API returned status %d with message: %s", resp.StatusCode, string(bodyBytes))
	}

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

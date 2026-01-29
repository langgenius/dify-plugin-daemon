package tool

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/langgenius/dify-plugin-daemon/cmd/dify_cli/types"
	"github.com/langgenius/dify-plugin-daemon/cmd/dify_cli/uploader"
	"github.com/langgenius/dify-plugin-daemon/pkg/utils/http_requests"
	"github.com/langgenius/dify-plugin-daemon/pkg/utils/parser"
)

func CallAPI(cfg *types.DifyConfig, tool *types.DifyToolDeclaration, params map[string]any, credentialID string) error {
	if cfg.Env.FilesURL != "" {
		SetFilesURL(cfg.Env.FilesURL)
	}

	if err := uploadFileParams(cfg, tool, params); err != nil {
		return err
	}

	reqBody := types.InvokeToolRequest{
		Type:           types.INVOKE_TYPE_TOOL,
		ToolType:       tool.ProviderType,
		Provider:       tool.Identity.Provider,
		Tool:           tool.Identity.Name,
		ToolParameters: params,
		CredentialId:   credentialID,
		CredentialType: tool.CredentialType,
	}

	body, err := json.Marshal(reqBody)
	if err != nil {
		return err
	}

	timestamp := strconv.FormatInt(time.Now().Unix(), 10)
	signature := uploader.SignRequest(cfg.Env.CliApiSecret, timestamp, body)
	url := strings.TrimSuffix(cfg.Env.CliApiURL, "/") + "/cli/api/invoke/tool"

	client := &http.Client{Timeout: 5 * time.Minute}
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

	return processResponse(resp.Body)
}

func uploadFileParams(cfg *types.DifyConfig, tool *types.DifyToolDeclaration, params map[string]any) error {
	for _, p := range tool.Parameters {
		value, exists := params[p.Name]
		if !exists {
			continue
		}

		switch p.Type {
		case types.ToolParameterTypeFile:
			path, ok := value.(string)
			if !ok {
				continue
			}
			fileObj, err := uploader.UploadFile(cfg, path)
			if err != nil {
				return fmt.Errorf("failed to upload file for parameter '%s': %w", p.Name, err)
			}
			params[p.Name] = fileObj

		case types.ToolParameterTypeFiles:
			pathStr, ok := value.(string)
			if !ok {
				continue
			}
			paths := strings.Split(pathStr, ",")
			var fileObjs []*types.ToolFileObject
			for _, path := range paths {
				path = strings.TrimSpace(path)
				fileObj, err := uploader.UploadFile(cfg, path)
				if err != nil {
					return fmt.Errorf("failed to upload file '%s' for parameter '%s': %w", path, p.Name, err)
				}
				fileObjs = append(fileObjs, fileObj)
			}
			params[p.Name] = fileObjs
		}
	}
	return nil
}

func processResponse(body io.Reader) error {
	return parser.LengthPrefixedChunking(body, 0x0f, 1024*1024*30, func(data []byte) error {
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

		return Dispatch(chunk.Data)
	})
}

func FetchToolsBatch(cfg *types.DifyConfig, refs []types.ToolReference) ([]types.DifyToolDeclaration, error) {
	if len(refs) == 0 {
		return nil, nil
	}

	items := make([]types.FetchToolItem, 0, len(refs))
	for _, ref := range refs {
		items = append(items, types.FetchToolItem{
			ToolType:     ref.ToolType,
			ToolProvider: ref.ToolProvider,
			ToolName:     ref.ToolName,
			CredentialID: ref.CredentialID,
		})
	}

	reqBody := types.FetchToolBatchRequest{Tools: items}
	body, err := json.Marshal(reqBody)
	if err != nil {
		return nil, err
	}

	timestamp := strconv.FormatInt(time.Now().Unix(), 10)
	signature := uploader.SignRequest(cfg.Env.CliApiSecret, timestamp, body)
	url := strings.TrimSuffix(cfg.Env.CliApiURL, "/") + "/cli/api/fetch/tools/batch"

	client := &http.Client{Timeout: 2 * time.Minute}
	req, err := http.NewRequest("POST", url, strings.NewReader(string(body)))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Cli-Api-Session-Id", cfg.Env.CliApiSessionID)
	req.Header.Set("X-Cli-Api-Timestamp", timestamp)
	req.Header.Set("X-Cli-Api-Signature", signature)

	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		bodyBytes, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("API returned status %d: %s", resp.StatusCode, string(bodyBytes))
	}

	var result types.DifyInnerAPIResponse[types.FetchToolBatchResponse]
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	if result.Error != "" {
		return nil, fmt.Errorf("API error: %s", result.Error)
	}

	if result.Data == nil {
		return nil, errors.New("response data is nil")
	}

	return result.Data.Tools, nil
}

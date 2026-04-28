package slim

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

const extractPath = "/v2/invoke/extract"

type daemonExtractResponse struct {
	Code    int           `json:"code"`
	Message string        `json:"message"`
	Data    ExtractResult `json:"data"`
}

func ExtractRemote(opts ExtractOptions, remote *RemoteConfig) (*ExtractResult, error) {
	if opts.Path != "" {
		return nil, NewError(ErrInvalidInput, "-path is not supported in remote mode")
	}
	if opts.PluginID == "" {
		return nil, NewError(ErrInvalidInput, "-id is required in remote mode")
	}

	client := NewDaemonClient(remote.DaemonAddr, remote.DaemonKey)
	return client.Extract(opts.PluginID)
}

func (c *DaemonClient) Extract(pluginID string) (*ExtractResult, error) {
	req, err := http.NewRequest(http.MethodGet, c.addr+extractPath, nil)
	if err != nil {
		return nil, NewError(ErrNetwork, err.Error())
	}
	req.Header.Set("X-Api-Key", c.key)
	req.Header.Set("X-Plugin-Unique-Identifier", pluginID)

	resp, err := c.client.Do(req)
	if err != nil {
		return nil, NewError(ErrNetwork, err.Error())
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, NewError(ErrNetwork, fmt.Sprintf("read daemon response: %s", err))
	}
	if resp.StatusCode != http.StatusOK {
		return nil, NewError(ErrDaemon, fmt.Sprintf("status %d: %s", resp.StatusCode, string(body)))
	}

	var parsed daemonExtractResponse
	if err := json.Unmarshal(body, &parsed); err != nil {
		return nil, NewError(ErrStreamParse, "invalid JSON response from daemon: "+err.Error())
	}
	if parsed.Code != 0 {
		return nil, NewError(ErrDaemon, parsed.Message)
	}

	return &parsed.Data, nil
}

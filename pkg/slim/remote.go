package slim

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

const dispatchPrefix = "/v2/invoke/dispatch"

type DaemonClient struct {
	addr   string
	key    string
	client *http.Client
}

func NewDaemonClient(addr, key string) *DaemonClient {
	return &DaemonClient{
		addr:   strings.TrimRight(addr, "/"),
		key:    key,
		client: &http.Client{Timeout: 600 * time.Second},
	}
}

type ActionRoute struct {
	Type string
	Path string
}

var ActionRoutes = map[string]ActionRoute{
	// tool
	"invoke_tool":                 {Type: "tool", Path: "/tool/invoke"},
	"validate_tool_credentials":   {Type: "tool", Path: "/tool/validate_credentials"},
	"get_tool_runtime_parameters": {Type: "tool", Path: "/tool/get_runtime_parameters"},

	// model
	"invoke_llm":                    {Type: "model", Path: "/llm/invoke"},
	"get_llm_num_tokens":            {Type: "model", Path: "/llm/num_tokens"},
	"invoke_text_embedding":         {Type: "model", Path: "/text_embedding/invoke"},
	"invoke_multimodal_embedding":   {Type: "model", Path: "/multimodal_embedding/invoke"},
	"get_text_embedding_num_tokens": {Type: "model", Path: "/text_embedding/num_tokens"},
	"invoke_rerank":                 {Type: "model", Path: "/rerank/invoke"},
	"invoke_multimodal_rerank":      {Type: "model", Path: "/multimodal_rerank/invoke"},
	"invoke_tts":                    {Type: "model", Path: "/tts/invoke"},
	"get_tts_model_voices":          {Type: "model", Path: "/tts/model/voices"},
	"invoke_speech2text":            {Type: "model", Path: "/speech2text/invoke"},
	"invoke_moderation":             {Type: "model", Path: "/moderation/invoke"},
	"validate_provider_credentials": {Type: "model", Path: "/model/validate_provider_credentials"},
	"validate_model_credentials":    {Type: "model", Path: "/model/validate_model_credentials"},
	"get_ai_model_schemas":          {Type: "model", Path: "/model/schema"},

	// agent strategy
	"invoke_agent_strategy": {Type: "agent_strategy", Path: "/agent_strategy/invoke"},

	// endpoint
	"invoke_endpoint": {Type: "endpoint", Path: "/endpoint/invoke"},

	// oauth
	"get_authorization_url": {Type: "oauth", Path: "/oauth/get_authorization_url"},
	"get_credentials":       {Type: "oauth", Path: "/oauth/get_credentials"},
	"refresh_credentials":   {Type: "oauth", Path: "/oauth/refresh_credentials"},

	// datasource
	"validate_datasource_credentials":                    {Type: "datasource", Path: "/datasource/validate_credentials"},
	"invoke_website_datasource_get_crawl":                {Type: "datasource", Path: "/datasource/get_website_crawl"},
	"invoke_online_document_datasource_get_pages":        {Type: "datasource", Path: "/datasource/get_online_document_pages"},
	"invoke_online_document_datasource_get_page_content": {Type: "datasource", Path: "/datasource/get_online_document_page_content"},
	"invoke_online_drive_browse_files":                   {Type: "datasource", Path: "/datasource/online_drive_browse_files"},
	"invoke_online_drive_download_file":                  {Type: "datasource", Path: "/datasource/online_drive_download_file"},

	// dynamic parameter
	"fetch_parameter_options": {Type: "dynamic_parameter", Path: "/dynamic_select/fetch_parameter_options"},

	// trigger
	"invoke_trigger_event":         {Type: "trigger", Path: "/trigger/invoke_event"},
	"dispatch_trigger_event":       {Type: "trigger", Path: "/trigger/dispatch_event"},
	"subscribe_trigger":            {Type: "trigger", Path: "/trigger/subscribe"},
	"unsubscribe_trigger":          {Type: "trigger", Path: "/trigger/unsubscribe"},
	"refresh_trigger":              {Type: "trigger", Path: "/trigger/refresh"},
	"validate_trigger_credentials": {Type: "trigger", Path: "/trigger/validate_credentials"},
}

func LookupRoute(action string) (ActionRoute, bool) {
	r, ok := ActionRoutes[action]
	return r, ok
}

func (c *DaemonClient) Dispatch(ctx *InvokeContext) (io.ReadCloser, error) {
	route, ok := LookupRoute(ctx.Action)
	if !ok {
		return nil, NewError(ErrUnknownAction, ctx.Action)
	}

	body, err := json.Marshal(ctx.Request)
	if err != nil {
		return nil, NewError(ErrInvalidInput, fmt.Sprintf("failed to marshal request: %s", err))
	}

	url := c.addr + dispatchPrefix + route.Path
	req, err := http.NewRequest(http.MethodPost, url, bytes.NewReader(body))
	if err != nil {
		return nil, NewError(ErrNetwork, err.Error())
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Api-Key", c.key)
	req.Header.Set("X-Plugin-Unique-Identifier", ctx.PluginID)

	resp, err := c.client.Do(req)
	if err != nil {
		return nil, NewError(ErrNetwork, err.Error())
	}

	if resp.StatusCode != http.StatusOK {
		defer resp.Body.Close()
		b, _ := io.ReadAll(resp.Body)
		return nil, NewError(ErrDaemon, fmt.Sprintf("status %d: %s", resp.StatusCode, string(b)))
	}

	return resp.Body, nil
}

func RunRemote(ctx *InvokeContext, remote *RemoteConfig, out *OutputWriter) error {
	client := NewDaemonClient(remote.DaemonAddr, remote.DaemonKey)

	body, err := client.Dispatch(ctx)
	if err != nil {
		return err
	}
	defer body.Close()

	scanner := bufio.NewScanner(body)
	for scanner.Scan() {
		line := scanner.Text()
		if !strings.HasPrefix(line, "data: ") {
			continue
		}
		payload := strings.TrimPrefix(line, "data: ")
		if !json.Valid([]byte(payload)) {
			out.Error(ErrStreamParse, "invalid JSON in SSE frame")
			return NewError(ErrStreamParse, "invalid JSON in SSE frame")
		}
		if err := out.Chunk(json.RawMessage(payload)); err != nil {
			return NewError(ErrStreamRead, err.Error())
		}
	}
	if err := scanner.Err(); err != nil {
		out.Error(ErrStreamRead, err.Error())
		return NewError(ErrStreamRead, err.Error())
	}

	out.Done()
	return nil
}

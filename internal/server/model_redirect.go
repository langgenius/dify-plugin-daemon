package server

import (
	"bytes"
	"encoding/json"
	"errors"
	"io"
	"mime"
	"net/http"
	"slices"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/langgenius/dify-plugin-daemon/internal/core/io_tunnel"
	"github.com/langgenius/dify-plugin-daemon/internal/core/io_tunnel/access_types"
	"github.com/langgenius/dify-plugin-daemon/internal/db"
	"github.com/langgenius/dify-plugin-daemon/internal/server/constants"
	"github.com/langgenius/dify-plugin-daemon/internal/server/controllers"
	"github.com/langgenius/dify-plugin-daemon/internal/service"
	"github.com/langgenius/dify-plugin-daemon/internal/types/app"
	"github.com/langgenius/dify-plugin-daemon/internal/types/models"
	"github.com/langgenius/dify-plugin-daemon/pkg/entities"
	"github.com/langgenius/dify-plugin-daemon/pkg/entities/model_entities"
	"github.com/langgenius/dify-plugin-daemon/pkg/entities/plugin_entities"
)

const (
	modelRedirectProtocol           = "model-redirect/v1"
	modelRedirectRequestKey         = "model_redirect_request"
	defaultMaxModelRedirectBodySize = 160 * 1024 * 1024
)

type modelRedirectRequest = plugin_entities.InvokePluginRequest[io_tunnel.ModelDispatchData]

type modelRedirectOperation struct {
	modelTypes []string
	invoke     func(*modelRedirectRequest, *gin.Context, int)
}

type modelRedirectDispatcher struct {
	operations   map[string]modelRedirectOperation
	installation func(pluginID, tenantID string) (*models.PluginInstallation, error)
	timeout      int
	maxBodyBytes int64
}

func streamModelRedirect[R any](action access_types.PluginAccessAction, bufferSize int, modelTypes ...string) modelRedirectOperation {
	return modelRedirectOperation{
		modelTypes: modelTypes,
		invoke: func(request *modelRedirectRequest, ctx *gin.Context, timeout int) {
			service.StreamModelRedirect[R](request, ctx, action, bufferSize, timeout)
		},
	}
}

func pollModelRedirect(action access_types.PluginAccessAction) modelRedirectOperation {
	return modelRedirectOperation{
		modelTypes: []string{"llm"},
		invoke: func(request *modelRedirectRequest, ctx *gin.Context, _ int) {
			service.PollModelRedirect(request, ctx, action)
		},
	}
}

func newModelRedirectDispatcher(config *app.Config) modelRedirectDispatcher {
	maxBodyBytes := config.MaxModelRedirectRequestBytes
	if maxBodyBytes <= 0 {
		maxBodyBytes = defaultMaxModelRedirectBodySize
	}
	return modelRedirectDispatcher{
		installation: lookupPluginInstallation,
		timeout:      config.PluginMaxExecutionTimeout,
		maxBodyBytes: maxBodyBytes,
		operations: map[string]modelRedirectOperation{
			"llm/invoke":                  streamModelRedirect[model_entities.LLMResultChunk](access_types.PLUGIN_ACCESS_ACTION_INVOKE_LLM, 512, "llm"),
			"llm/num_tokens":              streamModelRedirect[model_entities.LLMGetNumTokensResponse](access_types.PLUGIN_ACCESS_ACTION_GET_LLM_NUM_TOKENS, 1, "llm"),
			"text_embedding/invoke":       streamModelRedirect[model_entities.TextEmbeddingResult](access_types.PLUGIN_ACCESS_ACTION_INVOKE_TEXT_EMBEDDING, 1, "text-embedding"),
			"multimodal_embedding/invoke": streamModelRedirect[model_entities.MultimodalEmbeddingResult](access_types.PLUGIN_ACCESS_ACTION_INVOKE_MULTIMODAL_EMBEDDING, 1, "text-embedding", "multimodal-embedding"),
			"text_embedding/num_tokens":   streamModelRedirect[model_entities.GetTextEmbeddingNumTokensResponse](access_types.PLUGIN_ACCESS_ACTION_GET_TEXT_EMBEDDING_NUM_TOKENS, 1, "text-embedding"),
			"rerank/invoke":               streamModelRedirect[model_entities.RerankResult](access_types.PLUGIN_ACCESS_ACTION_INVOKE_RERANK, 1, "rerank"),
			"multimodal_rerank/invoke":    streamModelRedirect[model_entities.MultimodalRerankResult](access_types.PLUGIN_ACCESS_ACTION_INVOKE_MULTIMODAL_RERANK, 1, "rerank", "multimodal-rerank"),
			"tts/invoke":                  streamModelRedirect[model_entities.TTSResult](access_types.PLUGIN_ACCESS_ACTION_INVOKE_TTS, 512, "tts"),
			"tts/model/voices":            streamModelRedirect[model_entities.GetTTSVoicesResponse](access_types.PLUGIN_ACCESS_ACTION_GET_TTS_MODEL_VOICES, 1, "tts"),
			"speech2text/invoke":          streamModelRedirect[model_entities.Speech2TextResult](access_types.PLUGIN_ACCESS_ACTION_INVOKE_SPEECH2TEXT, 1, "speech2text"),
			"moderation/invoke":           streamModelRedirect[model_entities.ModerationResult](access_types.PLUGIN_ACCESS_ACTION_INVOKE_MODERATION, 1, "moderation"),
			"model/schema":                streamModelRedirect[model_entities.GetModelSchemasResponse](access_types.PLUGIN_ACCESS_ACTION_GET_AI_MODEL_SCHEMAS, 1, "llm", "text-embedding", "multimodal-embedding", "rerank", "multimodal-rerank", "tts", "speech2text", "moderation"),
			"model/polling/start":         pollModelRedirect(access_types.PLUGIN_ACCESS_ACTION_START_POLLING),
			"model/polling/check":         pollModelRedirect(access_types.PLUGIN_ACCESS_ACTION_CHECK_POLLING),
		},
	}
}

func (app *App) pluginModelRedirectGroup(group *gin.RouterGroup, config *app.Config) {
	dispatcher := newModelRedirectDispatcher(config)
	group.GET("/v1/capabilities", dispatcher.capabilities)
	group.POST("/:version/*operation", controllers.CollectActiveDispatchRequests(), dispatcher.resolve, app.RedirectPluginInvoke(), app.InitClusterID(), dispatcher.invoke)
}

func (d *modelRedirectDispatcher) capabilities(ctx *gin.Context) {
	operations := make([]string, 0, len(d.operations))
	for operation := range d.operations {
		operations = append(operations, operation)
	}
	slices.Sort(operations)
	ctx.JSON(http.StatusOK, entities.NewSuccessResponse(gin.H{
		"protocols":  []string{modelRedirectProtocol},
		"operations": operations,
	}))
}

func abortModelRedirect(ctx *gin.Context, status int, code string) {
	ctx.AbortWithStatusJSON(status, gin.H{"code": code, "message": code})
}

func (d *modelRedirectDispatcher) resolve(ctx *gin.Context) {
	operationName := strings.TrimPrefix(ctx.Param("operation"), "/")
	operation, supported := d.operations[operationName]
	if ctx.Param("version") != "v1" || !supported {
		abortModelRedirect(ctx, http.StatusNotFound, "redirect_protocol_unsupported")
		return
	}
	request, ok := decodeModelRedirect(ctx, operationName, operation, d.maxBodyBytes)
	if !ok {
		return
	}

	installation, err := d.installation(request.PluginID, request.TenantId)
	switch {
	case errors.Is(err, db.ErrDatabaseNotFound) || err == nil && installation == nil:
		abortModelRedirect(ctx, http.StatusNotFound, "redirect_target_not_installed")
	case err != nil:
		abortModelRedirect(ctx, http.StatusInternalServerError, "redirect_installation_lookup_failed")
	case installation.TenantID != request.TenantId:
		abortModelRedirect(ctx, http.StatusForbidden, "redirect_invalid_tenant")
	case installation.PluginID != request.PluginID:
		abortModelRedirect(ctx, http.StatusNotFound, "redirect_target_not_installed")
	case installation.PluginUniqueIdentifier != request.UniqueIdentifier.String():
		abortModelRedirect(ctx, http.StatusConflict, "redirect_target_version_mismatch")
	default:
		ctx.Set(constants.CONTEXT_KEY_PLUGIN_INSTALLATION, *installation)
		ctx.Set(constants.CONTEXT_KEY_PLUGIN_UNIQUE_IDENTIFIER, request.UniqueIdentifier)
		ctx.Set(modelRedirectRequestKey, request)
		ctx.Next()
	}
}

func (d *modelRedirectDispatcher) invoke(ctx *gin.Context) {
	value, _ := ctx.Get(modelRedirectRequestKey)
	request, ok := value.(*modelRedirectRequest)
	if !ok {
		abortModelRedirect(ctx, http.StatusInternalServerError, "redirect_invalid_request")
		return
	}
	d.operations[strings.TrimPrefix(ctx.Param("operation"), "/")].invoke(request, ctx, d.timeout)
}

func decodeModelRedirect(ctx *gin.Context, operationName string, operation modelRedirectOperation, maxBodyBytes int64) (*modelRedirectRequest, bool) {
	invalid := func() (*modelRedirectRequest, bool) {
		abortModelRedirect(ctx, http.StatusBadRequest, "redirect_invalid_request")
		return nil, false
	}
	mediaType, _, err := mime.ParseMediaType(ctx.GetHeader("Content-Type"))
	if err != nil || mediaType != "application/json" {
		return invalid()
	}
	if ctx.Request.ContentLength > maxBodyBytes {
		abortModelRedirect(ctx, http.StatusRequestEntityTooLarge, "redirect_request_too_large")
		return nil, false
	}
	body := http.MaxBytesReader(ctx.Writer, ctx.Request.Body, maxBodyBytes)
	content, err := io.ReadAll(body)
	var sizeErr *http.MaxBytesError
	if errors.As(err, &sizeErr) {
		abortModelRedirect(ctx, http.StatusRequestEntityTooLarge, "redirect_request_too_large")
		return nil, false
	}
	if err != nil {
		return invalid()
	}
	ctx.Request.Body = io.NopCloser(bytes.NewReader(content))

	envelope := redirectObject(content, []string{"user_id", "data", "redirect"}, "app_id")
	routing := redirectObject(envelope["redirect"], []string{"version", "route_id", "target"})
	target := redirectObject(routing["target"], []string{"plugin_id", "expected_unique_identifier", "provider", "model_type", "model", "credentials"})
	if envelope == nil || routing == nil || target == nil {
		return invalid()
	}
	var version int
	if json.Unmarshal(routing["version"], &version) != nil {
		return invalid()
	}
	if version != 1 {
		abortModelRedirect(ctx, http.StatusNotFound, "redirect_protocol_unsupported")
		return nil, false
	}
	routeID := redirectString(routing["route_id"])
	if id, err := uuid.Parse(routeID); err != nil || id.String() != routeID {
		return invalid()
	}
	tenantID := ctx.Param("tenant_id")
	if id, err := uuid.Parse(tenantID); err != nil || id.String() != tenantID {
		return invalid()
	}

	var userID string
	if json.Unmarshal(envelope["user_id"], &userID) != nil {
		return invalid()
	}
	var appID *string
	if raw, exists := envelope["app_id"]; exists && json.Unmarshal(raw, &appID) != nil {
		return invalid()
	}

	pluginID := redirectString(target["plugin_id"])
	identity, err := plugin_entities.NewPluginUniqueIdentifier(redirectString(target["expected_unique_identifier"]))
	if err != nil || identity.PluginID() != pluginID || redirectString(target["provider"]) == "" || redirectString(target["model"]) == "" {
		return invalid()
	}
	if len(uniqueRedirectObject(target["credentials"])) == 0 {
		return invalid()
	}
	data := uniqueRedirectObject(envelope["data"])
	modelType := redirectString(data["model_type"])
	if redirectString(data["provider"]) == "" || redirectString(data["model"]) == "" ||
		modelType != redirectString(target["model_type"]) || !slices.Contains(operation.modelTypes, modelType) {
		return invalid()
	}
	if sourceCredentials := uniqueRedirectObject(data["credentials"]); sourceCredentials == nil || len(sourceCredentials) != 0 {
		return invalid()
	}
	for name := range data {
		canonical := strings.ToLower(name)
		switch canonical {
		case "provider", "model", "model_type", "credentials", "credential_type", "tenant_id":
			if name != canonical {
				return invalid()
			}
		case "redirect", "model_redirect", "route_id", "target", "plugin_id", "unique_identifier", "plugin_unique_identifier", "expected_unique_identifier", "runtime_url",
			"type", "action", "context", "session_id", "user_id", "app_id", "conversation_id", "message_id", "endpoint_id",
			"authorization", "api_key", "access_token", "system_credentials", "credential_id":
			return invalid()
		}
	}
	if rawTenant, exists := data["tenant_id"]; exists && redirectString(rawTenant) != tenantID {
		abortModelRedirect(ctx, http.StatusForbidden, "redirect_invalid_tenant")
		return nil, false
	}
	if operationName == "tts/invoke" {
		data["tenant_id"], _ = json.Marshal(tenantID)
	}
	data["provider"] = target["provider"]
	data["model"] = target["model"]
	data["credentials"] = target["credentials"]
	delete(data, "credential_type")
	return &modelRedirectRequest{
		InvokePluginUserIdentity: plugin_entities.InvokePluginUserIdentity{TenantId: tenantID, UserId: userID},
		BasePluginIdentifier:     plugin_entities.BasePluginIdentifier{PluginID: pluginID},
		UniqueIdentifier:         identity,
		AppID:                    appID,
		Data:                     data,
	}, true
}

func redirectString(value json.RawMessage) string {
	var result string
	_ = json.Unmarshal(value, &result)
	return result
}

func redirectObject(raw json.RawMessage, required []string, optional ...string) io_tunnel.ModelDispatchData {
	fields := uniqueRedirectObject(raw)
	for _, name := range required {
		if _, exists := fields[name]; !exists {
			return nil
		}
	}
	for name, value := range fields {
		if (!slices.Contains(required, name) && !slices.Contains(optional, name)) || bytes.Equal(bytes.TrimSpace(value), []byte("null")) {
			return nil
		}
	}
	return fields
}

func uniqueRedirectObject(raw json.RawMessage) io_tunnel.ModelDispatchData {
	decoder := json.NewDecoder(bytes.NewReader(raw))
	token, err := decoder.Token()
	if err != nil || token != json.Delim('{') {
		return nil
	}
	fields := make(io_tunnel.ModelDispatchData)
	for decoder.More() {
		token, err := decoder.Token()
		name, isString := token.(string)
		if err != nil || !isString {
			return nil
		}
		if _, duplicate := fields[name]; duplicate {
			return nil
		}
		var value json.RawMessage
		if err := decoder.Decode(&value); err != nil {
			return nil
		}
		fields[name] = value
	}
	token, err = decoder.Token()
	if err != nil || token != json.Delim('}') || decoder.Decode(new(json.RawMessage)) != io.EOF {
		return nil
	}
	return fields
}

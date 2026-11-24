package server

import (
	"bytes"
	"encoding/json"
	"errors"
	"io"
	"mime"
	"net/http"
	"regexp"
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
	modelRedirectProtocolV1         = "model-redirect/v1"
	modelRedirectProtocolV2         = "model-redirect/v2"
	modelRedirectRequestKey         = "model_redirect_request"
	modelRedirectCurrentUIDKey      = "model_redirect_current_uid"
	modelRedirectPeerUIDHeader      = "X-Dify-Model-Redirect-Resolved-Uid"
	defaultMaxModelRedirectBodySize = 160 * 1024 * 1024
)

var modelRedirectPluginIDPattern = regexp.MustCompile(`^[a-z0-9_-]{1,64}/[a-z0-9_-]{1,128}$`)

type modelRedirectRequest = plugin_entities.InvokePluginRequest[io_tunnel.ModelDispatchData]

type modelRedirectOperation struct {
	modelTypes []string
	invoke     func(*modelRedirectRequest, *gin.Context, int)
}

type modelRedirectDispatcher struct {
	operations          map[string]modelRedirectOperation
	installation        func(pluginID, tenantID string) (*models.PluginInstallation, error)
	currentInstallation func(pluginID, tenantID string) (*models.PluginInstallation, error)
	verifyCurrentTarget func(plugin_entities.PluginUniqueIdentifier, string) bool
	timeout             int
	maxBodyBytes        int64
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
		installation:        lookupPluginInstallation,
		currentInstallation: getPluginInstallation,
		verifyCurrentTarget: func(uid plugin_entities.PluginUniqueIdentifier, provider string) bool {
			return service.VerifyCurrentModelRedirectTarget(config.MaxPluginPackageSize, uid, provider)
		},
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
		"protocols":  []string{modelRedirectProtocolV1, modelRedirectProtocolV2},
		"operations": operations,
	}))
}

func abortModelRedirect(ctx *gin.Context, status int, code string) {
	ctx.AbortWithStatusJSON(status, gin.H{"code": code, "message": code})
}

func (d *modelRedirectDispatcher) resolve(ctx *gin.Context) {
	operationName := strings.TrimPrefix(ctx.Param("operation"), "/")
	operation, supported := d.operations[operationName]
	var protocolVersion int
	switch ctx.Param("version") {
	case "v1":
		protocolVersion = 1
	case "v2":
		protocolVersion = 2
	}
	if protocolVersion == 0 || !supported {
		abortModelRedirect(ctx, http.StatusNotFound, "redirect_protocol_unsupported")
		return
	}
	request, ok := decodeModelRedirect(ctx, operationName, operation, d.maxBodyBytes, protocolVersion)
	if !ok {
		return
	}

	lookup := d.installation
	if protocolVersion == 2 {
		lookup = d.currentInstallation
	}
	installation, err := lookup(request.PluginID, request.TenantId)
	switch {
	case errors.Is(err, db.ErrDatabaseNotFound) || err == nil && installation == nil:
		abortModelRedirect(ctx, http.StatusNotFound, "redirect_target_not_installed")
	case err != nil:
		abortModelRedirect(ctx, http.StatusInternalServerError, "redirect_installation_lookup_failed")
	case installation.TenantID != request.TenantId:
		abortModelRedirect(ctx, http.StatusForbidden, "redirect_invalid_tenant")
	case installation.PluginID != request.PluginID:
		abortModelRedirect(ctx, http.StatusNotFound, "redirect_target_not_installed")
	case protocolVersion == 1 && installation.PluginUniqueIdentifier != request.UniqueIdentifier.String():
		abortModelRedirect(ctx, http.StatusConflict, "redirect_target_version_mismatch")
	default:
		if protocolVersion == 2 && !d.resolveCurrentTarget(ctx, request, installation) {
			return
		}
		ctx.Set(constants.CONTEXT_KEY_PLUGIN_INSTALLATION, *installation)
		ctx.Set(constants.CONTEXT_KEY_PLUGIN_UNIQUE_IDENTIFIER, request.UniqueIdentifier)
		ctx.Set(modelRedirectRequestKey, request)
		ctx.Next()
	}
}

func (d *modelRedirectDispatcher) resolveCurrentTarget(ctx *gin.Context, request *modelRedirectRequest, installation *models.PluginInstallation) bool {
	uid, err := plugin_entities.NewPluginUniqueIdentifier(installation.PluginUniqueIdentifier)
	runtimeType := plugin_entities.PluginRuntimeType(installation.RuntimeType)
	if err != nil || uid.PluginID() != request.PluginID ||
		(runtimeType != plugin_entities.PLUGIN_RUNTIME_TYPE_LOCAL && runtimeType != plugin_entities.PLUGIN_RUNTIME_TYPE_SERVERLESS) {
		abortModelRedirect(ctx, http.StatusForbidden, "redirect_target_untrusted")
		return false
	}
	peerUID := ctx.Request.Header.Values(modelRedirectPeerUIDHeader)
	if len(peerUID) > 1 || (len(peerUID) == 1 && peerUID[0] != uid.String()) {
		abortModelRedirect(ctx, http.StatusConflict, "redirect_target_version_mismatch")
		return false
	}
	if !d.verifyCurrentTarget(uid, redirectString(request.Data["provider"])) {
		abortModelRedirect(ctx, http.StatusForbidden, "redirect_target_untrusted")
		return false
	}
	request.UniqueIdentifier = uid
	ctx.Set(modelRedirectCurrentUIDKey, uid.String())
	ctx.Request = ctx.Request.WithContext(service.WithModelRedirectRuntimeType(ctx.Request.Context(), runtimeType))
	return true
}

func modelRedirectPeerRequest(ctx *gin.Context) *http.Request {
	uid := ctx.GetString(modelRedirectCurrentUIDKey)
	if uid == "" {
		return ctx.Request
	}
	request := ctx.Request.Clone(ctx.Request.Context())
	request.Header.Set(modelRedirectPeerUIDHeader, uid)
	return request
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

func decodeModelRedirect(ctx *gin.Context, operationName string, operation modelRedirectOperation, maxBodyBytes int64, protocolVersion int) (*modelRedirectRequest, bool) {
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
	requiredTarget := []string{"plugin_id", "provider", "model_type", "model", "credentials"}
	if protocolVersion == 1 {
		requiredTarget = append(requiredTarget, "expected_unique_identifier")
	}
	target := redirectObject(routing["target"], requiredTarget)
	if envelope == nil || routing == nil || target == nil {
		return invalid()
	}
	var version int
	if json.Unmarshal(routing["version"], &version) != nil {
		return invalid()
	}
	if version != protocolVersion {
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
	if redirectString(target["provider"]) == "" || redirectString(target["model"]) == "" {
		return invalid()
	}
	var identity plugin_entities.PluginUniqueIdentifier
	if protocolVersion == 1 {
		identity, err = plugin_entities.NewPluginUniqueIdentifier(redirectString(target["expected_unique_identifier"]))
		if err != nil || identity.PluginID() != pluginID {
			return invalid()
		}
	} else if !modelRedirectPluginIDPattern.MatchString(pluginID) {
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

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
	"github.com/langgenius/dify-plugin-daemon/internal/db"
	"github.com/langgenius/dify-plugin-daemon/internal/server/constants"
	"github.com/langgenius/dify-plugin-daemon/internal/server/controllers"
	"github.com/langgenius/dify-plugin-daemon/internal/service"
	"github.com/langgenius/dify-plugin-daemon/internal/types/app"
	"github.com/langgenius/dify-plugin-daemon/internal/types/models"
	"github.com/langgenius/dify-plugin-daemon/pkg/entities"
	"github.com/langgenius/dify-plugin-daemon/pkg/entities/plugin_entities"
	"github.com/langgenius/dify-plugin-daemon/pkg/entities/requests"
	"github.com/langgenius/dify-plugin-daemon/pkg/validators"
)

const modelRedirectProtocol = "model-redirect/v1"

// Routing instructions are request-local. They are never added to plugin Context
// or a session, where they could expose credentials to other plugins or storage.
type modelRedirectRequest struct {
	UserID   *string                    `json:"user_id"`
	AppID    *string                    `json:"app_id,omitempty"`
	Data     map[string]json.RawMessage `json:"data"`
	Redirect *modelRedirect             `json:"redirect"`
}

type modelRedirect struct {
	Version int                 `json:"version"`
	RouteID string              `json:"route_id"`
	Target  modelRedirectTarget `json:"target"`
}

type modelRedirectTarget struct {
	PluginID                 string         `json:"plugin_id"`
	ExpectedUniqueIdentifier *string        `json:"expected_unique_identifier,omitempty"`
	Provider                 string         `json:"provider"`
	ModelType                string         `json:"model_type"`
	Model                    string         `json:"model"`
	Credentials              map[string]any `json:"credentials"`
}

type modelRedirectOperation struct {
	modelTypes []string
	validate   func([]byte) error
	invoke     func(*gin.Context, *modelRedirectRequest, []byte)
}

type modelRedirectDispatcher struct {
	operations   map[string]modelRedirectOperation
	installation func(tenantID, pluginID string) (*models.PluginInstallation, error)
}

func redirectedOperation[T any](modelTypes []string, timeout int, invoke func(*plugin_entities.InvokePluginRequest[T], *gin.Context, int)) modelRedirectOperation {
	return modelRedirectOperation{
		modelTypes: modelTypes,
		validate: func(data []byte) error {
			var request T
			if err := json.Unmarshal(data, &request); err != nil {
				return err
			}
			return validators.GlobalEntitiesValidator.Struct(request)
		},
		invoke: func(ctx *gin.Context, envelope *modelRedirectRequest, data []byte) {
			var request T
			if err := json.Unmarshal(data, &request); err != nil {
				abortModelRedirect(ctx, http.StatusBadRequest, "redirect_invalid_request")
				return
			}
			identityValue, _ := ctx.Get(constants.CONTEXT_KEY_PLUGIN_UNIQUE_IDENTIFIER)
			identity, ok := identityValue.(plugin_entities.PluginUniqueIdentifier)
			if !ok {
				abortModelRedirect(ctx, http.StatusInternalServerError, "redirect_invalid_request")
				return
			}
			invoke(&plugin_entities.InvokePluginRequest[T]{
				InvokePluginUserIdentity: plugin_entities.InvokePluginUserIdentity{TenantId: ctx.Param("tenant_id"), UserId: *envelope.UserID},
				BasePluginIdentifier:     plugin_entities.BasePluginIdentifier{PluginID: envelope.Redirect.Target.PluginID},
				UniqueIdentifier:         identity,
				AppID:                    envelope.AppID,
				Data:                     request,
			}, ctx, timeout)
		},
	}
}

func newModelRedirectDispatcher(config *app.Config) *modelRedirectDispatcher {
	llm := []string{"llm"}
	embedding := []string{"text-embedding"}
	tts := []string{"tts"}
	all := []string{"llm", "text-embedding", "rerank", "tts", "speech2text", "moderation"}
	timeout := config.PluginMaxExecutionTimeout
	return &modelRedirectDispatcher{
		installation: func(tenantID, pluginID string) (*models.PluginInstallation, error) {
			return lookupPluginInstallation(pluginID, tenantID)
		},
		operations: map[string]modelRedirectOperation{
			"model/schema":                     redirectedOperation(all, timeout, service.GetAIModelSchema),
			"llm/invoke":                       redirectedOperation(llm, timeout, service.InvokeLLM),
			"llm/num_tokens":                   redirectedOperation(llm, timeout, service.GetLLMNumTokens),
			"text_embedding/invoke":            redirectedOperation(embedding, timeout, service.InvokeTextEmbedding),
			"text_embedding/num_tokens":        redirectedOperation(embedding, timeout, service.GetTextEmbeddingNumTokens),
			"rerank/invoke":                    redirectedOperation([]string{"rerank"}, timeout, service.InvokeRerank),
			"tts/invoke":                       redirectedOperation(tts, timeout, service.InvokeTTS),
			"tts/model/voices":                 redirectedOperation(tts, timeout, service.GetTTSModelVoices),
			"speech2text/invoke":               redirectedOperation([]string{"speech2text"}, timeout, service.InvokeSpeech2Text),
			"moderation/invoke":                redirectedOperation([]string{"moderation"}, timeout, service.InvokeModeration),
			"model/validate_model_credentials": redirectedOperation(all, timeout, service.ValidateModelCredentials),
			"model/polling/start": redirectedOperation(llm, timeout, func(r *plugin_entities.InvokePluginRequest[requests.RequestStartPolling], c *gin.Context, _ int) {
				service.StartPolling(r, c)
			}),
			"model/polling/check": redirectedOperation(llm, timeout, func(r *plugin_entities.InvokePluginRequest[requests.RequestCheckPolling], c *gin.Context, _ int) {
				service.CheckPolling(r, c)
			}),
		},
	}
}

func (app *App) pluginModelRedirectGroup(group *gin.RouterGroup, config *app.Config) {
	dispatcher := newModelRedirectDispatcher(config)
	group.GET("/v1/capabilities", dispatcher.capabilities)
	group.POST("/:version/*operation",
		controllers.CollectActiveDispatchRequests(),
		dispatcher.resolve,
		app.RedirectPluginInvoke(),
		app.InitClusterID(),
		dispatcher.invoke,
	)
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
	// Do not include decode/validation errors: they can contain credential values.
	ctx.AbortWithStatusJSON(status, gin.H{"code": code, "message": code})
}

func (d *modelRedirectDispatcher) resolve(ctx *gin.Context) {
	if ctx.Param("version") != "v1" {
		abortModelRedirect(ctx, http.StatusNotFound, "redirect_protocol_unsupported")
		return
	}
	operation, supported := d.operations[strings.TrimPrefix(ctx.Param("operation"), "/")]
	if !supported {
		abortModelRedirect(ctx, http.StatusNotFound, "redirect_protocol_unsupported")
		return
	}
	envelope, data, ok := decodeModelRedirect(ctx, operation)
	if !ok {
		return
	}
	if err := operation.validate(data); err != nil {
		abortModelRedirect(ctx, http.StatusBadRequest, "redirect_invalid_request")
		return
	}
	target := envelope.Redirect.Target
	installation, err := d.installation(ctx.Param("tenant_id"), target.PluginID)
	if errors.Is(err, db.ErrDatabaseNotFound) || (err == nil && installation == nil) {
		abortModelRedirect(ctx, http.StatusNotFound, "redirect_target_not_installed")
		return
	}
	if err != nil {
		abortModelRedirect(ctx, http.StatusInternalServerError, "redirect_invalid_request")
		return
	}
	if installation.TenantID != ctx.Param("tenant_id") {
		abortModelRedirect(ctx, http.StatusForbidden, "redirect_invalid_tenant")
		return
	}
	identity, err := plugin_entities.NewPluginUniqueIdentifier(installation.PluginUniqueIdentifier)
	if err != nil || identity.PluginID() != target.PluginID || installation.PluginID != target.PluginID {
		abortModelRedirect(ctx, http.StatusNotFound, "redirect_target_not_installed")
		return
	}
	if target.ExpectedUniqueIdentifier != nil && *target.ExpectedUniqueIdentifier != identity.String() {
		abortModelRedirect(ctx, http.StatusConflict, "redirect_target_version_mismatch")
		return
	}
	ctx.Request.Header.Set(constants.X_PLUGIN_ID, target.PluginID)
	ctx.Set(constants.CONTEXT_KEY_PLUGIN_INSTALLATION, *installation)
	ctx.Set(constants.CONTEXT_KEY_PLUGIN_UNIQUE_IDENTIFIER, identity)
	ctx.Next()
}

func (d *modelRedirectDispatcher) invoke(ctx *gin.Context) {
	operation := d.operations[strings.TrimPrefix(ctx.Param("operation"), "/")]
	// Decode again after cluster routing. Keeping the original body replayable
	// lets a remote node resolve the same tenant installation/version itself,
	// without persisting credentials in a Gin context or trusting a runtime URL.
	envelope, data, ok := decodeModelRedirect(ctx, operation)
	if ok {
		operation.invoke(ctx, envelope, data)
	}
}

// encoding/json matches struct field names case-insensitively and accepts null
// for pointers. V1's JSON contract permits neither alternate keys nor null for
// optional strings, so check the exact object keys before typed decoding.
func redirectObject(raw []byte, required, optional []string) (map[string]json.RawMessage, bool) {
	var fields map[string]json.RawMessage
	if json.Unmarshal(raw, &fields) != nil || fields == nil {
		return nil, false
	}
	for _, name := range required {
		if _, exists := fields[name]; !exists {
			return nil, false
		}
	}
	for name, value := range fields {
		if !slices.Contains(required, name) && !slices.Contains(optional, name) {
			return nil, false
		}
		if bytes.Equal(bytes.TrimSpace(value), []byte("null")) {
			return nil, false
		}
	}
	return fields, true
}

func decodeModelRedirect(ctx *gin.Context, operation modelRedirectOperation) (*modelRedirectRequest, []byte, bool) {
	fail := func(status int, code string) (*modelRedirectRequest, []byte, bool) {
		abortModelRedirect(ctx, status, code)
		return nil, nil, false
	}
	mediaType, _, err := mime.ParseMediaType(ctx.GetHeader("Content-Type"))
	if err != nil || mediaType != "application/json" {
		return fail(http.StatusBadRequest, "redirect_invalid_request")
	}
	content, err := io.ReadAll(ctx.Request.Body)
	if err != nil {
		return fail(http.StatusBadRequest, "redirect_invalid_request")
	}
	ctx.Request.Body = io.NopCloser(bytes.NewReader(content))
	fields, valid := redirectObject(content, []string{"user_id", "data", "redirect"}, []string{"app_id"})
	if !valid {
		return fail(http.StatusBadRequest, "redirect_invalid_request")
	}
	routingFields, valid := redirectObject(fields["redirect"], []string{"version", "route_id", "target"}, nil)
	if !valid {
		return fail(http.StatusBadRequest, "redirect_invalid_request")
	}
	if _, valid := redirectObject(routingFields["target"], []string{"plugin_id", "provider", "model_type", "model", "credentials"}, []string{"expected_unique_identifier"}); !valid {
		return fail(http.StatusBadRequest, "redirect_invalid_request")
	}
	decoder := json.NewDecoder(bytes.NewReader(content))
	decoder.DisallowUnknownFields()
	var envelope modelRedirectRequest
	if err := decoder.Decode(&envelope); err != nil || decoder.Decode(new(any)) != io.EOF {
		return fail(http.StatusBadRequest, "redirect_invalid_request")
	}
	if envelope.UserID == nil || envelope.Redirect == nil || envelope.Data == nil {
		return fail(http.StatusBadRequest, "redirect_invalid_request")
	}
	if envelope.Redirect.Version != 1 {
		return fail(http.StatusNotFound, "redirect_protocol_unsupported")
	}
	routeID, err := uuid.Parse(envelope.Redirect.RouteID)
	if err != nil || routeID.String() != envelope.Redirect.RouteID {
		return fail(http.StatusBadRequest, "redirect_invalid_request")
	}
	target := envelope.Redirect.Target
	if target.PluginID == "" || target.Provider == "" || target.Model == "" || len(target.Credentials) == 0 ||
		(target.ExpectedUniqueIdentifier != nil && *target.ExpectedUniqueIdentifier == "") {
		return fail(http.StatusBadRequest, "redirect_invalid_request")
	}
	var source struct {
		Provider    string                     `json:"provider"`
		Model       string                     `json:"model"`
		ModelType   string                     `json:"model_type"`
		Credentials map[string]json.RawMessage `json:"credentials"`
	}
	for _, name := range []string{"provider", "model", "model_type", "credentials"} {
		if _, exists := envelope.Data[name]; !exists {
			return fail(http.StatusBadRequest, "redirect_invalid_request")
		}
	}
	for name := range envelope.Data {
		for _, reserved := range []string{"provider", "model", "model_type", "credentials", "credential_type", "tenant_id", "redirect", "model_redirect", "context", "runtime_url", "unique_identifier", "plugin_id"} {
			if name != reserved && strings.EqualFold(name, reserved) {
				return fail(http.StatusBadRequest, "redirect_invalid_request")
			}
		}
	}
	sourceData, err := json.Marshal(envelope.Data)
	if err != nil || json.Unmarshal(sourceData, &source) != nil || source.Provider == "" || source.Model == "" ||
		source.Credentials == nil || len(source.Credentials) != 0 || source.ModelType != target.ModelType ||
		!slices.Contains(operation.modelTypes, target.ModelType) {
		return fail(http.StatusBadRequest, "redirect_invalid_request")
	}
	// The operation data is extensible, but routing control fields must never be
	// nested there. A TTS tenant field is accepted only for the URI's tenant.
	for _, forbidden := range []string{"redirect", "model_redirect", "context", "runtime_url", "unique_identifier", "plugin_id"} {
		if _, exists := envelope.Data[forbidden]; exists {
			return fail(http.StatusBadRequest, "redirect_invalid_request")
		}
	}
	if rawTenant, exists := envelope.Data["tenant_id"]; exists {
		var tenantID string
		if json.Unmarshal(rawTenant, &tenantID) != nil || tenantID != ctx.Param("tenant_id") {
			return fail(http.StatusForbidden, "redirect_invalid_tenant")
		}
	}
	if target.ModelType == "tts" {
		envelope.Data["tenant_id"], _ = json.Marshal(ctx.Param("tenant_id"))
	}
	// Reuse the existing operation validation before and after replacement. No
	// caller-provided source authentication hint is carried to the destination.
	sourceData, _ = json.Marshal(envelope.Data)
	if operation.validate(sourceData) != nil {
		return fail(http.StatusBadRequest, "redirect_invalid_request")
	}
	envelope.Data["provider"], _ = json.Marshal(target.Provider)
	envelope.Data["model"], _ = json.Marshal(target.Model)
	envelope.Data["credentials"], _ = json.Marshal(target.Credentials)
	delete(envelope.Data, "credential_type")
	data, _ := json.Marshal(envelope.Data)
	return &envelope, data, true
}

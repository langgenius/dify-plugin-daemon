package server

import (
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/langgenius/dify-plugin-daemon/internal/db"
	"github.com/langgenius/dify-plugin-daemon/internal/server/constants"
	"github.com/langgenius/dify-plugin-daemon/internal/types/app"
	"github.com/langgenius/dify-plugin-daemon/internal/types/models"
	"github.com/langgenius/dify-plugin-daemon/pkg/entities/plugin_entities"
	"github.com/stretchr/testify/require"
)

const (
	redirectTenantID  = "e6c5e5fb-a7af-47a7-b67b-f627a458a282"
	redirectTargetID  = "author/destination"
	redirectTargetUID = "author/destination:1.0.0@0123456789abcdef0123456789abcdef"
	redirectTestURL   = "/plugin/" + redirectTenantID + "/dispatch/redirect/"
	redirectTestKey   = "test-daemon-api-key"
)

const redirectTestBody = `{
    "user_id":"trusted-user", "app_id":"original-app",
    "data":{
        "provider":"source", "model":"source-model", "model_type":"llm", "credentials":{}, "credential_type":"api-key",
        "model_parameters":{"future_option":{"count":9007199254740993,"ratio":1.00000000000000000003}},
        "prompt_messages":[{"role":"user","content":[{"type":"image","future_image":{"bytes":"AA+/"}}],"opaque_body":{"future":null}}],
        "tools":[{"unique_tool_field":true}],"json_schema":{"future_schema":"<raw>"},"stop":["stop"],
        "stream":true,"future_parameter":{"provider":"keep","tenant_id":"keep","credentials":"model argument"},"explicit_null":null
    },
    "redirect":{"version":1,"route_id":"267af2c2-1ebc-414e-a5ba-adab58970b87","target":{
        "plugin_id":"` + redirectTargetID + `","expected_unique_identifier":"` + redirectTargetUID + `",
        "provider":"target","model":"target-model","model_type":"llm","credentials":{"api_key":"test-managed-key"}
    }}
}`

func testRedirectRouter(dispatcher modelRedirectDispatcher) *gin.Engine {
	router := gin.New()
	group := router.Group("/plugin/:tenant_id/dispatch/redirect", CheckingKey(redirectTestKey))
	group.GET("/v1/capabilities", dispatcher.capabilities)
	group.POST("/:version/*operation", dispatcher.resolve, dispatcher.invoke)
	return router
}

func postTestRedirect(router http.Handler, url, body, token, contentType string) *httptest.ResponseRecorder {
	req := httptest.NewRequest(http.MethodPost, url, strings.NewReader(body))
	req.Header.Set("Content-Type", contentType)
	req.Header.Set(constants.X_API_KEY, token)
	req.Header.Set(constants.X_PLUGIN_ID, "author/source")
	req.Header.Set(constants.PluginUniqueIdentifier, "arbitrary-header-ignored")
	resp := httptest.NewRecorder()
	router.ServeHTTP(resp, req)
	return resp
}

func TestModelRedirectPreservesOperationData(t *testing.T) {
	d := newModelRedirectDispatcher(&app.Config{PluginMaxExecutionTimeout: 7})
	d.installation = func(pluginID, tenantID string) (*models.PluginInstallation, error) {
		require.Equal(t, redirectTargetID, pluginID)
		require.Equal(t, redirectTenantID, tenantID)
		return &models.PluginInstallation{TenantID: tenantID, PluginID: pluginID, PluginUniqueIdentifier: redirectTargetUID}, nil
	}
	var original struct {
		Data map[string]json.RawMessage `json:"data"`
	}
	require.NoError(t, json.Unmarshal([]byte(redirectTestBody), &original))
	called := 0
	d.operations["llm/invoke"] = modelRedirectOperation{
		modelTypes: []string{"llm"},
		invoke: func(req *modelRedirectRequest, ctx *gin.Context, timeout int) {
			called++
			require.Equal(t, 7, timeout)
			require.Equal(t, "trusted-user", req.UserId)
			require.NotNil(t, req.AppID)
			require.Equal(t, "original-app", *req.AppID)
			require.Equal(t, redirectTargetUID, req.UniqueIdentifier.String())
			require.Equal(t, redirectTenantID, req.TenantId)
			require.Nil(t, req.Context)
			require.Equal(t, `"target"`, string(req.Data["provider"]))
			require.Equal(t, `"target-model"`, string(req.Data["model"]))
			require.Equal(t, `{"api_key":"test-managed-key"}`, string(req.Data["credentials"]))
			require.NotContains(t, req.Data, "credential_type")
			for key, value := range original.Data {
				switch key {
				case "provider", "model", "credentials", "credential_type":
				default:
					require.Equal(t, value, req.Data[key], key)
				}
			}
			replayed, err := io.ReadAll(ctx.Request.Body)
			require.NoError(t, err)
			require.Equal(t, redirectTestBody, string(replayed), "peer must receive the same authenticated envelope")
			ctx.Status(http.StatusNoContent)
		},
	}
	resp := postTestRedirect(testRedirectRouter(d), redirectTestURL+"v1/llm/invoke", redirectTestBody, redirectTestKey, "application/json; charset=utf-8")
	require.Equal(t, http.StatusNoContent, resp.Code, resp.Body.String())
	require.Equal(t, 1, called)
}

func TestModelRedirectRejectsInvalidBoundary(t *testing.T) {
	extraData := func(field string) string {
		return strings.Replace(redirectTestBody, `"data":{`, `"data":{`+field+`,`, 1)
	}
	const contentType = "application/json"
	for _, tt := range []struct {
		name, operation, body, token, mediaType string
		status                                  int
	}{
		{"auth", "v1/llm/invoke", redirectTestBody, "wrong", contentType, 401},
		{"operation", "v1/tool/invoke", redirectTestBody, redirectTestKey, contentType, 404},
		{"credential validation is not usage", "v1/model/validate_model_credentials", redirectTestBody, redirectTestKey, contentType, 404},
		{"version", "v3/llm/invoke", redirectTestBody, redirectTestKey, contentType, 404},
		{"non-JSON", "v1/llm/invoke", redirectTestBody, redirectTestKey, "text/plain", 400},
		{"duplicate envelope", "v1/llm/invoke", strings.Replace(redirectTestBody, `"user_id":"trusted-user"`, `"user_id":"other","user_id":"trusted-user"`, 1), redirectTestKey, contentType, 400},
		{"trailing content", "v1/llm/invoke", redirectTestBody + `{}`, redirectTestKey, contentType, 400},
		{"nonempty source credentials", "v1/llm/invoke", strings.Replace(redirectTestBody, `"credentials":{}`, `"credentials":{"api_key":"source-secret"}`, 1), redirectTestKey, contentType, 400},
		{"null source credentials", "v1/llm/invoke", strings.Replace(redirectTestBody, `"credentials":{}`, `"credentials":null`, 1), redirectTestKey, contentType, 400},
		{"duplicate credentials", "v1/llm/invoke", extraData(`"credentials":{"api_key":"source-secret"}`), redirectTestKey, contentType, 400},
		{"case alias", "v1/llm/invoke", extraData(`"Credentials":{"api_key":"source-secret"}`), redirectTestKey, contentType, 400},
		{"action smuggling", "v1/llm/invoke", extraData(`"action":"invoke_tool"`), redirectTestKey, contentType, 400},
		{"session context smuggling", "v1/llm/invoke", extraData(`"Context":{"user_id":"other"}`), redirectTestKey, contentType, 400},
		{"source auth alias", "v1/llm/invoke", extraData(`"AUTHORIZATION":"source-secret"`), redirectTestKey, contentType, 400},
		{"runtime URL smuggling", "v1/llm/invoke", extraData(`"runtime_url":"https://example.invalid"`), redirectTestKey, contentType, 400},
		{"cross tenant", "v1/llm/invoke", extraData(`"tenant_id":"another-tenant"`), redirectTestKey, contentType, 403},
		{"model type mismatch", "v1/rerank/invoke", redirectTestBody, redirectTestKey, contentType, 400},
		{"missing pinned UID", "v1/llm/invoke", strings.Replace(redirectTestBody, `,"expected_unique_identifier":"`+redirectTargetUID+`"`, "", 1), redirectTestKey, contentType, 400},
		{"bad plugin ID", "v1/llm/invoke", strings.Replace(redirectTestBody, `"plugin_id":"`+redirectTargetID+`"`, `"plugin_id":"author/other"`, 1), redirectTestKey, contentType, 400},
	} {
		t.Run(tt.name, func(t *testing.T) {
			d := newModelRedirectDispatcher(&app.Config{})
			d.installation = func(string, string) (*models.PluginInstallation, error) {
				t.Fatal("invalid request reached installation lookup")
				return nil, nil
			}
			resp := postTestRedirect(testRedirectRouter(d), redirectTestURL+tt.operation, tt.body, tt.token, tt.mediaType)
			require.Equal(t, tt.status, resp.Code, resp.Body.String())
			require.NotContains(t, resp.Body.String(), "source-secret")
			require.NotContains(t, resp.Body.String(), "test-managed-key")
		})
	}
}

func currentRedirectTestBody() string {
	body := strings.Replace(redirectTestBody, `,"expected_unique_identifier":"`+redirectTargetUID+`"`, "", 1)
	return strings.Replace(body, `"version":1`, `"version":2`, 1)
}

func TestModelRedirectFollowsCurrentVerifiedInstallation(t *testing.T) {
	d := newModelRedirectDispatcher(&app.Config{})
	d.installation = func(string, string) (*models.PluginInstallation, error) {
		t.Fatal("version-independent routing must not use the v1 installation cache")
		return nil, nil
	}
	secondUID := strings.Replace(redirectTargetUID, ":1.0.0@", ":2.0.0@", 1)
	currentUID := redirectTargetUID
	d.currentInstallation = func(pluginID, tenantID string) (*models.PluginInstallation, error) {
		require.Equal(t, redirectTargetID, pluginID)
		require.Equal(t, redirectTenantID, tenantID)
		return &models.PluginInstallation{TenantID: tenantID, PluginID: pluginID, PluginUniqueIdentifier: currentUID, RuntimeType: "local"}, nil
	}
	var verified []string
	d.verifyCurrentTarget = func(uid plugin_entities.PluginUniqueIdentifier, provider string) bool {
		require.Equal(t, currentUID, uid.String())
		require.Equal(t, "target", provider)
		verified = append(verified, uid.String())
		return true
	}
	body := currentRedirectTestBody()
	var invoked []string
	op := d.operations["llm/invoke"]
	op.invoke = func(req *modelRedirectRequest, ctx *gin.Context, _ int) {
		invoked = append(invoked, req.UniqueIdentifier.String())
		require.Equal(t, currentUID, req.UniqueIdentifier.String())
		require.Equal(t, req.UniqueIdentifier, ctx.MustGet(constants.CONTEXT_KEY_PLUGIN_UNIQUE_IDENTIFIER))
		require.Equal(t, redirectTenantID, req.TenantId)
		require.Equal(t, `{"api_key":"test-managed-key"}`, string(req.Data["credentials"]))
		require.Equal(t, currentUID, modelRedirectPeerRequest(ctx).Header.Get(modelRedirectPeerUIDHeader))
		require.Empty(t, ctx.Request.Header.Get(modelRedirectPeerUIDHeader))
		raw, err := io.ReadAll(ctx.Request.Body)
		require.NoError(t, err)
		require.Equal(t, body, string(raw))
		ctx.Status(http.StatusNoContent)
	}
	d.operations["llm/invoke"] = op
	router := testRedirectRouter(d)
	for _, uid := range []string{redirectTargetUID, secondUID} {
		currentUID = uid
		response := postTestRedirect(router, redirectTestURL+"v2/llm/invoke", body, redirectTestKey, "application/json")
		require.Equal(t, http.StatusNoContent, response.Code, response.Body.String())
	}
	require.Equal(t, []string{redirectTargetUID, secondUID}, verified)
	require.Equal(t, verified, invoked)
}

func TestModelRedirectCurrentRejectsUntrustedInstallation(t *testing.T) {
	installation := models.PluginInstallation{TenantID: redirectTenantID, PluginID: redirectTargetID, PluginUniqueIdentifier: redirectTargetUID, RuntimeType: "local"}
	for _, tt := range []struct {
		name, tenant, plugin, uid, runtime string
		peerUID, apiKey, body              string
		allowVerification                  bool
		status                             int
	}{
		{name: "unauthenticated", apiKey: "wrong", status: 401},
		{name: "cross tenant", tenant: "other-tenant", status: 403},
		{name: "wrong plugin", plugin: "other/target", status: 404},
		{name: "invalid stored UID", uid: "not-an-identifier", status: 403},
		{name: "UID belongs to another plugin", uid: strings.Replace(redirectTargetUID, "author/", "other/", 1), status: 403},
		{name: "debug runtime", runtime: "remote", status: 403},
		{name: "unverified package", allowVerification: true, status: 403},
		{name: "peer observed an upgrade", peerUID: strings.Replace(redirectTargetUID, ":1.0.0@", ":2.0.0@", 1), status: 409},
		{name: "client UID in body", body: strings.Replace(redirectTestBody, `"version":1`, `"version":2`, 1), status: 400},
		{name: "plugin ID containing UID", body: strings.Replace(currentRedirectTestBody(), `"plugin_id":"`+redirectTargetID+`"`, `"plugin_id":"`+redirectTargetUID+`"`, 1), status: 400},
	} {
		t.Run(tt.name, func(t *testing.T) {
			d := newModelRedirectDispatcher(&app.Config{})
			d.currentInstallation = func(pluginID, tenantID string) (*models.PluginInstallation, error) {
				if tt.status == 400 || tt.status == 401 {
					t.Fatal("invalid request reached tenant lookup")
				}
				require.Equal(t, redirectTargetID, pluginID)
				require.Equal(t, redirectTenantID, tenantID)
				result := installation
				if tt.tenant != "" {
					result.TenantID = tt.tenant
				}
				if tt.plugin != "" {
					result.PluginID = tt.plugin
				}
				if tt.uid != "" {
					result.PluginUniqueIdentifier = tt.uid
				}
				if tt.runtime != "" {
					result.RuntimeType = tt.runtime
				}
				return &result, nil
			}
			d.verifyCurrentTarget = func(uid plugin_entities.PluginUniqueIdentifier, provider string) bool {
				require.True(t, tt.allowVerification, "invalid request reached package verification")
				require.Equal(t, redirectTargetUID, uid.String())
				require.Equal(t, "target", provider)
				return false
			}
			op := d.operations["llm/invoke"]
			op.invoke = func(*modelRedirectRequest, *gin.Context, int) { t.Fatal("untrusted request reached runtime") }
			d.operations["llm/invoke"] = op
			body := tt.body
			if body == "" {
				body = currentRedirectTestBody()
			}
			token := tt.apiKey
			if token == "" {
				token = redirectTestKey
			}
			req := httptest.NewRequest(http.MethodPost, redirectTestURL+"v2/llm/invoke", strings.NewReader(body))
			req.Header.Set("Content-Type", "application/json")
			req.Header.Set(constants.X_API_KEY, token)
			if tt.peerUID != "" {
				req.Header.Set(modelRedirectPeerUIDHeader, tt.peerUID)
			}
			resp := httptest.NewRecorder()
			testRedirectRouter(d).ServeHTTP(resp, req)
			require.Equal(t, tt.status, resp.Code, resp.Body.String())
			require.NotContains(t, resp.Body.String(), "test-managed-key")
		})
	}
}

func TestModelRedirectRequestSizeLimit(t *testing.T) {
	for _, length := range []int64{int64(len(redirectTestBody)), -1} {
		d := newModelRedirectDispatcher(&app.Config{MaxModelRedirectRequestBytes: 16})
		d.installation = func(string, string) (*models.PluginInstallation, error) {
			t.Error("oversized request reached installation lookup")
			return nil, nil
		}
		req := httptest.NewRequest(http.MethodPost, redirectTestURL+"v1/llm/invoke", strings.NewReader(redirectTestBody))
		req.ContentLength = length
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set(constants.X_API_KEY, redirectTestKey)
		resp := httptest.NewRecorder()
		testRedirectRouter(d).ServeHTTP(resp, req)
		require.Equal(t, http.StatusRequestEntityTooLarge, resp.Code, resp.Body.String())
		require.NotContains(t, resp.Body.String(), "test-managed-key")
	}
}

func TestModelRedirectRequiresTenantInstallationAndExactUID(t *testing.T) {
	installation := models.PluginInstallation{TenantID: redirectTenantID, PluginID: redirectTargetID, PluginUniqueIdentifier: redirectTargetUID}
	wrongTenant := installation
	wrongTenant.TenantID = "other-tenant"
	wrongUID := installation
	wrongUID.PluginUniqueIdentifier = strings.ReplaceAll(redirectTargetUID, ":1.0.0@", ":2.0.0@")
	wrongPlugin := installation
	wrongPlugin.PluginID = "author/other"
	for _, tt := range []struct {
		name         string
		installation *models.PluginInstallation
		err          error
		status       int
	}{
		{"missing", nil, db.ErrDatabaseNotFound, 404},
		{"wrong tenant", &wrongTenant, nil, 403},
		{"wrong plugin", &wrongPlugin, nil, 404},
		{"replaced version", &wrongUID, nil, 409},
	} {
		t.Run(tt.name, func(t *testing.T) {
			d := newModelRedirectDispatcher(&app.Config{})
			d.installation = func(string, string) (*models.PluginInstallation, error) { return tt.installation, tt.err }
			operation := d.operations["llm/invoke"]
			operation.invoke = func(*modelRedirectRequest, *gin.Context, int) { t.Fatal("invalid installation reached runtime") }
			d.operations["llm/invoke"] = operation
			resp := postTestRedirect(testRedirectRouter(d), redirectTestURL+"v1/llm/invoke", redirectTestBody, redirectTestKey, "application/json")
			require.Equal(t, tt.status, resp.Code, resp.Body.String())
		})
	}
}

func TestModelRedirectBindsTTSInvocationTenant(t *testing.T) {
	d := newModelRedirectDispatcher(&app.Config{})
	d.installation = func(pluginID, tenantID string) (*models.PluginInstallation, error) {
		return &models.PluginInstallation{TenantID: tenantID, PluginID: pluginID, PluginUniqueIdentifier: redirectTargetUID}, nil
	}
	operation := d.operations["tts/invoke"]
	operation.invoke = func(req *modelRedirectRequest, ctx *gin.Context, _ int) {
		require.Equal(t, `"`+redirectTenantID+`"`, string(req.Data["tenant_id"]))
		require.Equal(t, `"source-voice"`, string(req.Data["voice"]))
		ctx.Status(http.StatusNoContent)
	}
	d.operations["tts/invoke"] = operation
	body := strings.ReplaceAll(redirectTestBody, `"model_type":"llm"`, `"model_type":"tts"`)
	body = strings.Replace(body, `"data":{`, `"data":{"voice":"source-voice",`, 1)
	resp := postTestRedirect(testRedirectRouter(d), redirectTestURL+"v1/tts/invoke", body, redirectTestKey, "application/json")
	require.Equal(t, http.StatusNoContent, resp.Code, resp.Body.String())
}

func TestModelRedirectCapabilities(t *testing.T) {
	router := gin.New()
	pluginGroup := router.Group("/plugin/:tenant_id")
	(&App{}).pluginGroup(pluginGroup, &app.Config{ServerKey: redirectTestKey})
	req := httptest.NewRequest(http.MethodGet, redirectTestURL+"v1/capabilities", nil)
	resp := httptest.NewRecorder()
	router.ServeHTTP(resp, req)
	require.Equal(t, http.StatusUnauthorized, resp.Code)
	req.Header.Set(constants.X_API_KEY, redirectTestKey)
	resp = httptest.NewRecorder()
	router.ServeHTTP(resp, req)
	require.Equal(t, http.StatusOK, resp.Code, resp.Body.String())
	var result struct {
		Data struct {
			Protocols  []string `json:"protocols"`
			Operations []string `json:"operations"`
		} `json:"data"`
	}
	require.NoError(t, json.Unmarshal(resp.Body.Bytes(), &result))
	require.Equal(t, []string{"model-redirect/v1", "model-redirect/v2"}, result.Data.Protocols)
	require.Equal(t, []string{
		"llm/invoke", "llm/num_tokens", "model/polling/check", "model/polling/start", "model/schema", "moderation/invoke",
		"multimodal_embedding/invoke", "multimodal_rerank/invoke", "rerank/invoke", "speech2text/invoke", "text_embedding/invoke",
		"text_embedding/num_tokens", "tts/invoke", "tts/model/voices",
	}, result.Data.Operations)
}

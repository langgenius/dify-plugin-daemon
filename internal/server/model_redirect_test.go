package server

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/langgenius/dify-plugin-daemon/internal/db"
	"github.com/langgenius/dify-plugin-daemon/internal/types/app"
	"github.com/langgenius/dify-plugin-daemon/internal/types/models"
	"github.com/langgenius/dify-plugin-daemon/pkg/entities/plugin_entities"
	"github.com/langgenius/dify-plugin-daemon/pkg/entities/requests"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const redirectTestIdentity = "example/target:1.0.0@aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa"
const redirectTestSecret = "TEST_ONLY_DO_NOT_LOG"

func redirectTestPayload() map[string]any {
	return map[string]any{
		"user_id": "user-1", "app_id": "app-1",
		"data": map[string]any{
			"provider": "source", "model": "source-model", "model_type": "llm",
			"credentials": map[string]any{}, "stream": true,
			"prompt_messages": []any{map[string]any{"role": "user", "content": "hello"}},
		},
		"redirect": map[string]any{
			"version": 1, "route_id": "33333333-3333-4333-8333-333333333333",
			"target": map[string]any{
				"plugin_id": "example/target", "provider": "target", "model": "target-model", "model_type": "llm",
				"credentials": map[string]any{"api_key": redirectTestSecret, "endpoint_url": "https://model.example/v1"},
			},
		},
	}
}

func redirectTestTarget(payload map[string]any) map[string]any {
	return payload["redirect"].(map[string]any)["target"].(map[string]any)
}

func redirectTestDispatcher(t *testing.T) *modelRedirectDispatcher {
	t.Helper()
	dispatcher := newModelRedirectDispatcher(&app.Config{PluginMaxExecutionTimeout: 10})
	dispatcher.installation = func(tenantID, pluginID string) (*models.PluginInstallation, error) {
		require.Equal(t, "tenant-1", tenantID)
		require.Equal(t, "example/target", pluginID)
		return &models.PluginInstallation{TenantID: tenantID, PluginID: pluginID, PluginUniqueIdentifier: redirectTestIdentity}, nil
	}
	return dispatcher
}

func redirectTestRouter(dispatcher *modelRedirectDispatcher, between ...gin.HandlerFunc) *gin.Engine {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	group := router.Group("/plugin/:tenant_id/dispatch/redirect", CheckingKey("test-key"))
	group.GET("/v1/capabilities", dispatcher.capabilities)
	handlers := []gin.HandlerFunc{dispatcher.resolve}
	handlers = append(handlers, between...)
	handlers = append(handlers, dispatcher.invoke)
	group.POST("/:version/*operation", handlers...)
	return router
}

func requestRedirect(t *testing.T, router http.Handler, operation string, payload map[string]any) *httptest.ResponseRecorder {
	t.Helper()
	body, err := json.Marshal(payload)
	require.NoError(t, err)
	request := httptest.NewRequest(http.MethodPost, "/plugin/tenant-1/dispatch/redirect/v1/"+operation, bytes.NewReader(body))
	request.Header.Set("Content-Type", "application/json")
	request.Header.Set("X-Api-Key", "test-key")
	// This source need not be installed. The target is resolved from the route.
	request.Header.Set("X-Plugin-ID", "example/source")
	response := httptest.NewRecorder()
	router.ServeHTTP(response, request)
	return response
}

func TestModelRedirectDispatchesTypedTargetOnly(t *testing.T) {
	dispatcher := redirectTestDispatcher(t)
	calls := 0
	dispatcher.operations["llm/invoke"] = redirectedOperation([]string{"llm"}, 10,
		func(request *plugin_entities.InvokePluginRequest[requests.RequestInvokeLLM], ctx *gin.Context, timeout int) {
			calls++
			assert.Equal(t, "tenant-1", request.TenantId)
			assert.Equal(t, "user-1", request.UserId)
			require.NotNil(t, request.AppID)
			assert.Equal(t, "app-1", *request.AppID)
			assert.Equal(t, "example/target", request.PluginID)
			assert.Equal(t, redirectTestIdentity, request.UniqueIdentifier.String())
			assert.Equal(t, "target", request.Data.Provider)
			assert.Equal(t, "target-model", request.Data.Model)
			assert.Equal(t, redirectTestSecret, request.Data.Credentials.Credentials["api_key"])
			assert.Equal(t, "https://model.example/v1", request.Data.Credentials.Credentials["endpoint_url"])
			assert.Empty(t, request.Data.CredentialType)
			assert.Nil(t, request.Context)
			assert.True(t, request.Data.Stream)
			assert.Len(t, request.Data.Tools, 1)
			assert.Equal(t, []string{"END"}, request.Data.Stop)
			assert.Equal(t, 0.25, request.Data.ModelParameters["temperature"])
			assert.Equal(t, "object", request.Data.JSONSchema["type"])
			assert.Equal(t, 10, timeout)
			ctx.Header("Content-Type", "text/event-stream")
			ctx.String(http.StatusOK, "data: {\"code\":0,\"data\":{\"delta\":\"ok\"}}\n\n")
		})
	payload := redirectTestPayload()
	data := payload["data"].(map[string]any)
	data["credential_type"] = "source-oauth"
	data["model_parameters"] = map[string]any{"temperature": 0.25}
	data["tools"] = []any{map[string]any{"name": "weather", "parameters": map[string]any{"type": "object"}}}
	data["stop"] = []string{"END"}
	data["json_schema"] = map[string]any{"type": "object"}
	redirectTestTarget(payload)["expected_unique_identifier"] = redirectTestIdentity
	router := redirectTestRouter(dispatcher, func(ctx *gin.Context) {
		assert.Equal(t, "example/target", ctx.GetHeader("X-Plugin-ID"))
		assert.NotContains(t, fmt.Sprint(ctx.Keys), redirectTestSecret)
		ctx.Next()
	})
	for range 2 {
		response := requestRedirect(t, router, "llm/invoke", payload)
		assert.Equal(t, http.StatusOK, response.Code)
		assert.Contains(t, response.Body.String(), "\"delta\":\"ok\"")
		assert.NotContains(t, response.Body.String(), redirectTestSecret)
	}
	// Route IDs are correlation metadata, not cross-request idempotency keys.
	assert.Equal(t, 2, calls)
}

func TestModelRedirectRejectsInvalidRequestsBeforeInstallation(t *testing.T) {
	cases := []struct {
		name   string
		modify func(map[string]any)
		code   string
		status int
	}{
		{"source key", func(p map[string]any) {
			p["data"].(map[string]any)["credentials"] = map[string]any{"api_key": redirectTestSecret}
		}, "redirect_invalid_request", 400},
		{"missing source credentials", func(p map[string]any) { delete(p["data"].(map[string]any), "credentials") }, "redirect_invalid_request", 400},
		{"null source credentials", func(p map[string]any) { p["data"].(map[string]any)["credentials"] = nil }, "redirect_invalid_request", 400},
		{"missing user", func(p map[string]any) { delete(p, "user_id") }, "redirect_invalid_request", 400},
		{"noncanonical envelope key", func(p map[string]any) { p["User_ID"] = p["user_id"]; delete(p, "user_id") }, "redirect_invalid_request", 400},
		{"noncanonical target key", func(p map[string]any) { redirectTestTarget(p)["PROVIDER"] = "other" }, "redirect_invalid_request", 400},
		{"noncanonical credential hint", func(p map[string]any) { p["data"].(map[string]any)["CREDENTIAL_TYPE"] = "source-oauth" }, "redirect_invalid_request", 400},
		{"null app", func(p map[string]any) { p["app_id"] = nil }, "redirect_invalid_request", 400},
		{"missing route", func(p map[string]any) { delete(p, "redirect") }, "redirect_invalid_request", 400},
		{"context rejected", func(p map[string]any) { p["context"] = map[string]any{"secret": redirectTestSecret} }, "redirect_invalid_request", 400},
		{"route ID is canonical", func(p map[string]any) {
			p["redirect"].(map[string]any)["route_id"] = "AAAAAAAA-AAAA-AAAA-AAAA-AAAAAAAAAAAA"
		}, "redirect_invalid_request", 400},
		{"target tenant forbidden", func(p map[string]any) { redirectTestTarget(p)["tenant_id"] = "tenant-2" }, "redirect_invalid_request", 400},
		{"runtime URL forbidden", func(p map[string]any) { redirectTestTarget(p)["runtime_url"] = "https://other.example" }, "redirect_invalid_request", 400},
		{"recursive target forbidden", func(p map[string]any) { redirectTestTarget(p)["redirect"] = map[string]any{} }, "redirect_invalid_request", 400},
		{"recursive data forbidden", func(p map[string]any) { p["data"].(map[string]any)["redirect"] = map[string]any{} }, "redirect_invalid_request", 400},
		{"empty target credentials", func(p map[string]any) { redirectTestTarget(p)["credentials"] = map[string]any{} }, "redirect_invalid_request", 400},
		{"null expected version", func(p map[string]any) { redirectTestTarget(p)["expected_unique_identifier"] = nil }, "redirect_invalid_request", 400},
		{"source target type mismatch", func(p map[string]any) { redirectTestTarget(p)["model_type"] = "text-embedding" }, "redirect_invalid_request", 400},
		{"operation type mismatch", func(p map[string]any) {
			p["data"].(map[string]any)["model_type"] = "rerank"
			redirectTestTarget(p)["model_type"] = "rerank"
		}, "redirect_invalid_request", 400},
		{"typed data invalid", func(p map[string]any) { p["data"].(map[string]any)["stream"] = "not-a-bool" }, "redirect_invalid_request", 400},
		{"typed tools invalid", func(p map[string]any) { p["data"].(map[string]any)["tools"] = []any{map[string]any{"name": ""}} }, "redirect_invalid_request", 400},
		{"forged nested tenant", func(p map[string]any) { p["data"].(map[string]any)["tenant_id"] = "tenant-2" }, "redirect_invalid_tenant", 403},
		{"unsupported version", func(p map[string]any) { p["redirect"].(map[string]any)["version"] = 2 }, "redirect_protocol_unsupported", 404},
	}
	for _, test := range cases {
		t.Run(test.name, func(t *testing.T) {
			dispatcher := redirectTestDispatcher(t)
			dispatcher.installation = func(string, string) (*models.PluginInstallation, error) {
				t.Fatal("invalid request must not resolve or invoke a plugin")
				return nil, nil
			}
			payload := redirectTestPayload()
			test.modify(payload)
			response := requestRedirect(t, redirectTestRouter(dispatcher), "llm/invoke", payload)
			assert.Equal(t, test.status, response.Code)
			assert.Contains(t, response.Body.String(), test.code)
			assert.NotContains(t, response.Body.String(), redirectTestSecret)
		})
	}
}

func TestModelRedirectInstallationBoundary(t *testing.T) {
	cases := []struct {
		name         string
		installation *models.PluginInstallation
		err          error
		expected     string
		code         string
		status       int
	}{
		{"missing", nil, db.ErrDatabaseNotFound, "", "redirect_target_not_installed", 404},
		{"failed lookup", nil, errors.New(redirectTestSecret), "", "redirect_invalid_request", 500},
		{"other tenant", &models.PluginInstallation{TenantID: "tenant-2"}, nil, "", "redirect_invalid_tenant", 403},
		{"wrong plugin", &models.PluginInstallation{TenantID: "tenant-1", PluginID: "example/other", PluginUniqueIdentifier: redirectTestIdentity}, nil, "", "redirect_target_not_installed", 404},
		{"wrong version", &models.PluginInstallation{TenantID: "tenant-1", PluginID: "example/target", PluginUniqueIdentifier: redirectTestIdentity}, nil, "example/target:2.0.0@bbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbb", "redirect_target_version_mismatch", 409},
	}
	for _, test := range cases {
		t.Run(test.name, func(t *testing.T) {
			dispatcher := redirectTestDispatcher(t)
			dispatcher.installation = func(tenantID, pluginID string) (*models.PluginInstallation, error) {
				assert.Equal(t, "tenant-1", tenantID)
				assert.Equal(t, "example/target", pluginID)
				return test.installation, test.err
			}
			payload := redirectTestPayload()
			if test.expected != "" {
				redirectTestTarget(payload)["expected_unique_identifier"] = test.expected
			}
			response := requestRedirect(t, redirectTestRouter(dispatcher), "llm/invoke", payload)
			assert.Equal(t, test.status, response.Code)
			assert.Contains(t, response.Body.String(), test.code)
			assert.NotContains(t, response.Body.String(), redirectTestSecret)
		})
	}
}

func TestModelRedirectAdvertisedTypedOperations(t *testing.T) {
	cases := []struct {
		operation string
		modelType string
		data      map[string]any
	}{
		{"llm/invoke", "llm", map[string]any{"stream": false}},
		{"llm/num_tokens", "llm", nil},
		{"model/schema", "llm", nil},
		{"model/validate_model_credentials", "llm", nil},
		{"text_embedding/invoke", "text-embedding", map[string]any{"texts": []string{"hello"}, "input_type": "document"}},
		{"text_embedding/num_tokens", "text-embedding", map[string]any{"texts": []string{"hello"}}},
		{"rerank/invoke", "rerank", map[string]any{"query": "hello", "docs": []string{"doc"}}},
		{"tts/invoke", "tts", map[string]any{"content_text": "hello", "voice": "voice"}},
		{"tts/model/voices", "tts", nil},
		{"speech2text/invoke", "speech2text", map[string]any{"file": "aabb"}},
		{"moderation/invoke", "moderation", map[string]any{"text": "hello"}},
		{"model/polling/start", "llm", map[string]any{"stream": false}},
		{"model/polling/check", "llm", map[string]any{"plugin_state": map[string]any{"task_id": "fixture"}}},
	}
	dispatcher := redirectTestDispatcher(t)
	require.Len(t, dispatcher.operations, len(cases))
	for _, test := range cases {
		t.Run(test.operation, func(t *testing.T) {
			payload := redirectTestPayload()
			data := payload["data"].(map[string]any)
			data["model_type"] = test.modelType
			redirectTestTarget(payload)["model_type"] = test.modelType
			for key, value := range test.data {
				data[key] = value
			}
			operation := dispatcher.operations[test.operation]
			calls := 0
			operation.invoke = func(ctx *gin.Context, _ *modelRedirectRequest, normalized []byte) {
				calls++
				assert.NoError(t, operation.validate(normalized))
				assert.Contains(t, string(normalized), `"model":"target-model"`)
				if test.modelType == "tts" {
					assert.Contains(t, string(normalized), `"tenant_id":"tenant-1"`)
				}
				ctx.Status(http.StatusNoContent)
			}
			dispatcher.operations[test.operation] = operation
			response := requestRedirect(t, redirectTestRouter(dispatcher), test.operation, payload)
			assert.Equal(t, http.StatusNoContent, response.Code, response.Body.String())
			assert.Equal(t, 1, calls)
		})
	}
}

func TestModelRedirectCapabilitiesAndAuth(t *testing.T) {
	dispatcher := redirectTestDispatcher(t)
	router := redirectTestRouter(dispatcher)
	path := "/plugin/tenant-1/dispatch/redirect/v1/capabilities"
	for _, authenticated := range []bool{false, true} {
		request := httptest.NewRequest(http.MethodGet, path, nil)
		if authenticated {
			request.Header.Set("X-Api-Key", "test-key")
		}
		response := httptest.NewRecorder()
		router.ServeHTTP(response, request)
		if !authenticated {
			assert.Equal(t, http.StatusUnauthorized, response.Code)
			continue
		}
		var body struct {
			Data struct {
				Protocols  []string `json:"protocols"`
				Operations []string `json:"operations"`
			} `json:"data"`
		}
		require.NoError(t, json.Unmarshal(response.Body.Bytes(), &body))
		assert.Equal(t, []string{modelRedirectProtocol}, body.Data.Protocols)
		assert.Len(t, body.Data.Operations, len(dispatcher.operations))
	}
	for _, operation := range []string{"unknown/invoke", "tool/invoke", "multimodal_embedding/invoke"} {
		response := requestRedirect(t, router, operation, redirectTestPayload())
		assert.Equal(t, http.StatusNotFound, response.Code)
		assert.Contains(t, response.Body.String(), "redirect_protocol_unsupported")
	}
}

func TestModelRedirectCrossNodeStreamAndCancellation(t *testing.T) {
	var calls atomic.Int32
	cancelled := make(chan struct{})
	target := redirectTestDispatcher(t)
	operation := target.operations["llm/invoke"]
	operation.invoke = func(ctx *gin.Context, _ *modelRedirectRequest, _ []byte) {
		calls.Add(1)
		ctx.Header("Content-Type", "text/event-stream")
		ctx.String(http.StatusOK, "data: {\"code\":0,\"data\":{\"delta\":\"first\"}}\n\n")
		ctx.Writer.Flush()
		<-ctx.Request.Context().Done()
		close(cancelled)
	}
	target.operations["llm/invoke"] = operation
	targetServer := httptest.NewServer(redirectTestRouter(target))
	defer targetServer.Close()
	entry := redirectTestDispatcher(t)
	entryRouter := redirectTestRouter(entry, func(ctx *gin.Context) {
		// The cluster sees the original envelope with its body restored and the
		// target installation already resolved; the target repeats that check.
		request, err := http.NewRequestWithContext(ctx.Request.Context(), http.MethodPost, targetServer.URL+ctx.Request.URL.Path, ctx.Request.Body)
		require.NoError(t, err)
		request.Header = ctx.Request.Header.Clone()
		response, err := http.DefaultClient.Do(request)
		require.NoError(t, err)
		copyPluginRedirectResponse(ctx, response.StatusCode, response.Header, response.Body)
		ctx.Abort()
	})
	entryServer := httptest.NewServer(entryRouter)
	defer entryServer.Close()
	body, err := json.Marshal(redirectTestPayload())
	require.NoError(t, err)
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	request, err := http.NewRequestWithContext(ctx, http.MethodPost, entryServer.URL+"/plugin/tenant-1/dispatch/redirect/v1/llm/invoke", bytes.NewReader(body))
	require.NoError(t, err)
	request.Header.Set("Content-Type", "application/json")
	request.Header.Set("X-Api-Key", "test-key")
	response, err := http.DefaultClient.Do(request)
	require.NoError(t, err)
	defer response.Body.Close()
	assert.Equal(t, "text/event-stream", response.Header.Get("Content-Type"))
	line, err := bufio.NewReader(response.Body).ReadString('\n')
	require.NoError(t, err)
	assert.Contains(t, line, "first")
	cancel()
	select {
	case <-cancelled:
	case <-time.After(3 * time.Second):
		t.Fatal("target stream did not receive cancellation")
	}
	assert.Equal(t, int32(1), calls.Load())
}

func TestModelRedirectMalformedJSONAndNoFallback(t *testing.T) {
	dispatcher := redirectTestDispatcher(t)
	dispatcher.installation = func(string, string) (*models.PluginInstallation, error) {
		t.Fatal("malformed JSON must be rejected before resolving a plugin")
		return nil, nil
	}
	router := redirectTestRouter(dispatcher)
	for _, content := range []string{`{`, `null`, `{}` + `{}`} {
		request := httptest.NewRequest(http.MethodPost, "/plugin/tenant-1/dispatch/redirect/v1/llm/invoke", strings.NewReader(content))
		request.Header.Set("Content-Type", "application/json")
		request.Header.Set("X-Api-Key", "test-key")
		response := httptest.NewRecorder()
		router.ServeHTTP(response, request)
		assert.Equal(t, http.StatusBadRequest, response.Code)
	}
	request := httptest.NewRequest(http.MethodPost, "/plugin/tenant-1/dispatch/redirect/v2/llm/invoke", nil)
	request.Header.Set("X-Api-Key", "test-key")
	response := httptest.NewRecorder()
	router.ServeHTTP(response, request)
	assert.Equal(t, http.StatusNotFound, response.Code)
	assert.Contains(t, response.Body.String(), "redirect_protocol_unsupported")
}

func TestCopyPluginRedirectResponsePreservesErrorFrame(t *testing.T) {
	gin.SetMode(gin.TestMode)
	response := httptest.NewRecorder()
	ctx, _ := gin.CreateTestContext(response)
	frame := "data: {\"code\":-500,\"message\":\"target error\",\"data\":null}\n\n"
	copyPluginRedirectResponse(ctx, http.StatusOK, http.Header{"Content-Type": []string{"text/event-stream"}}, io.NopCloser(strings.NewReader(frame)))
	assert.Equal(t, frame, response.Body.String())
	assert.True(t, response.Flushed)
}

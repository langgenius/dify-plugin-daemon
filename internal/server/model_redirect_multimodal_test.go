package server

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/langgenius/dify-plugin-daemon/internal/types/models"
	"github.com/langgenius/dify-plugin-daemon/pkg/entities/plugin_entities"
	"github.com/langgenius/dify-plugin-daemon/pkg/entities/requests"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type multimodalRelayBody struct {
	io.ReadCloser
	closed func()
	once   sync.Once
}

func (body *multimodalRelayBody) Close() error {
	err := body.ReadCloser.Close()
	body.once.Do(body.closed)
	return err
}

func multimodalTestRelay(t *testing.T, target string, closed func()) gin.HandlerFunc {
	t.Helper()
	client := &http.Client{
		Transport: &http.Transport{Proxy: nil},
		Timeout:   5 * time.Second,
		CheckRedirect: func(*http.Request, []*http.Request) error {
			return http.ErrUseLastResponse
		},
	}
	t.Cleanup(client.CloseIdleConnections)
	return func(ctx *gin.Context) {
		request, err := http.NewRequestWithContext(ctx.Request.Context(), http.MethodPost, target+ctx.Request.URL.Path, ctx.Request.Body)
		if !assert.NoError(t, err) {
			ctx.AbortWithStatus(http.StatusBadGateway)
			return
		}
		request.Header = ctx.Request.Header.Clone()
		response, err := client.Do(request)
		if err != nil {
			if ctx.Request.Context().Err() == nil {
				t.Errorf("local target relay failed: %v", err)
			}
			ctx.Abort()
			return
		}
		if closed != nil {
			response.Body = &multimodalRelayBody{ReadCloser: response.Body, closed: closed}
		}
		copyPluginRedirectResponse(ctx, response.StatusCode, response.Header, response.Body)
		ctx.Abort()
	}
}

func multimodalObservedTarget(t *testing.T, router http.Handler, bodies chan<- []byte) *httptest.Server {
	t.Helper()
	server := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		body, err := io.ReadAll(request.Body)
		if !assert.NoError(t, err) {
			writer.WriteHeader(http.StatusBadRequest)
			return
		}
		assert.Equal(t, "test-key", request.Header.Get("X-Api-Key"))
		assert.Equal(t, "example/target", request.Header.Get("X-Plugin-ID"))
		assert.Equal(t, "original-body-relay", request.Header.Get("X-Redirect-Test"))
		request.Body = io.NopCloser(bytes.NewReader(body))
		select {
		case bodies <- body:
		default:
			t.Error("unexpected repeated target request")
			writer.WriteHeader(http.StatusBadGateway)
			return
		}
		router.ServeHTTP(writer, request)
	}))
	t.Cleanup(server.Close)
	return server
}

func multimodalEntryRequest(t *testing.T, ctx context.Context, url string, inputType string) (*http.Request, []byte) {
	t.Helper()
	payload := multimodalRedirectPayload(inputType)
	payload["data"].(map[string]any)["credential_type"] = "original-oauth"
	body, err := json.MarshalIndent(payload, "", "  ")
	require.NoError(t, err)
	request, err := http.NewRequestWithContext(ctx, http.MethodPost, url+"/plugin/tenant-1/dispatch/redirect/v1/multimodal_embedding/invoke", bytes.NewReader(body))
	require.NoError(t, err)
	request.Header.Set("Content-Type", "application/json")
	request.Header.Set("X-Api-Key", "test-key")
	request.Header.Set("X-Plugin-ID", "example/source")
	request.Header.Set("X-Redirect-Test", "original-body-relay")
	return request, body
}

func assertMultimodalTarget(t *testing.T, request *plugin_entities.InvokePluginRequest[requests.RequestInvokeMultimodalEmbedding], inputType string) {
	t.Helper()
	assert.Equal(t, "tenant-1", request.TenantId)
	assert.Equal(t, "user-1", request.UserId)
	if assert.NotNil(t, request.AppID) {
		assert.Equal(t, "app-1", *request.AppID)
	}
	assert.Equal(t, "example/target", request.PluginID)
	assert.Equal(t, redirectTestIdentity, request.UniqueIdentifier.String())
	assert.Equal(t, "target", request.Data.Provider)
	assert.Equal(t, "target-model", request.Data.Model)
	assert.Equal(t, "text-embedding", request.Data.ModelType)
	assert.Equal(t, inputType, request.Data.InputType)
	if assert.Len(t, request.Data.Documents, 2) {
		assert.Equal(t, "unchanged text", request.Data.Documents[0].Content)
		assert.Equal(t, "text", string(request.Data.Documents[0].ContentType))
		assert.Equal(t, "data:image/png;base64,AQID", request.Data.Documents[1].Content)
		assert.Equal(t, "image", string(request.Data.Documents[1].ContentType))
	}
	assert.Equal(t, redirectTestSecret, request.Data.Credentials.Credentials["api_key"])
	assert.Empty(t, request.Data.CredentialType)
	assert.Nil(t, request.Context)
}

func multimodalRedirectPayload(inputType string) map[string]any {
	payload := redirectTestPayload()
	data := payload["data"].(map[string]any)
	delete(data, "stream")
	delete(data, "prompt_messages")
	data["model_type"] = "text-embedding"
	data["input_type"] = inputType
	data["documents"] = []any{
		map[string]any{"content": "unchanged text", "content_type": "text"},
		map[string]any{"content": "data:image/png;base64,AQID", "content_type": "image"},
	}
	redirectTestTarget(payload)["model_type"] = "text-embedding"
	redirectTestTarget(payload)["expected_unique_identifier"] = redirectTestIdentity
	return payload
}

func TestMultimodalRedirectPreservesTypedDocumentsAndIdentity(t *testing.T) {
	for _, inputType := range []string{"document", "query"} {
		t.Run(inputType, func(t *testing.T) {
			dispatcher := redirectTestDispatcher(t)
			calls := 0
			dispatcher.operations["multimodal_embedding/invoke"] = redirectedOperation([]string{"text-embedding"}, 10,
				func(request *plugin_entities.InvokePluginRequest[requests.RequestInvokeMultimodalEmbedding], ctx *gin.Context, timeout int) {
					calls++
					assert.Equal(t, "tenant-1", request.TenantId)
					assert.Equal(t, "user-1", request.UserId)
					require.NotNil(t, request.AppID)
					assert.Equal(t, "app-1", *request.AppID)
					assert.Equal(t, "example/target", request.PluginID)
					assert.Equal(t, redirectTestIdentity, request.UniqueIdentifier.String())
					assert.Equal(t, "target", request.Data.Provider)
					assert.Equal(t, "target-model", request.Data.Model)
					assert.Equal(t, "text-embedding", request.Data.ModelType)
					assert.Equal(t, inputType, request.Data.InputType)
					require.Len(t, request.Data.Documents, 2)
					assert.Equal(t, "unchanged text", request.Data.Documents[0].Content)
					assert.Equal(t, "text", string(request.Data.Documents[0].ContentType))
					assert.Equal(t, "data:image/png;base64,AQID", request.Data.Documents[1].Content)
					assert.Equal(t, "image", string(request.Data.Documents[1].ContentType))
					assert.Equal(t, redirectTestSecret, request.Data.Credentials.Credentials["api_key"])
					assert.Empty(t, request.Data.CredentialType)
					assert.Nil(t, request.Context)
					assert.Equal(t, 10, timeout)
					ctx.JSON(http.StatusOK, gin.H{"embeddings": [][]float64{{1, 2}, {3, 4}}})
				})
			payload := multimodalRedirectPayload(inputType)
			payload["data"].(map[string]any)["credential_type"] = "original-oauth"
			response := requestRedirect(t, redirectTestRouter(dispatcher), "multimodal_embedding/invoke", payload)
			assert.Equal(t, http.StatusOK, response.Code)
			assert.Equal(t, 1, calls)
			assert.NotContains(t, response.Body.String(), redirectTestSecret)
		})
	}
}

func TestMultimodalRedirectRejectsMalformedOrWrongTypeBeforeLookup(t *testing.T) {
	cases := map[string]func(map[string]any){
		"wrong type": func(p map[string]any) {
			p["data"].(map[string]any)["model_type"] = "llm"
			redirectTestTarget(p)["model_type"] = "llm"
		},
		"missing documents":  func(p map[string]any) { delete(p["data"].(map[string]any), "documents") },
		"unknown input type": func(p map[string]any) { p["data"].(map[string]any)["input_type"] = "search" },
		"invalid document": func(p map[string]any) {
			p["data"].(map[string]any)["documents"] = []any{map[string]any{"content": "", "content_type": "image"}}
		},
		"unknown content type": func(p map[string]any) {
			p["data"].(map[string]any)["documents"] = []any{map[string]any{"content": "hello", "content_type": "video"}}
		},
		"source credentials": func(p map[string]any) {
			p["data"].(map[string]any)["credentials"] = map[string]any{"api_key": redirectTestSecret}
		},
	}
	for name, modify := range cases {
		t.Run(name, func(t *testing.T) {
			dispatcher := redirectTestDispatcher(t)
			dispatcher.installation = func(string, string) (*models.PluginInstallation, error) {
				t.Fatal("invalid multimodal request must not resolve a target")
				return nil, nil
			}
			payload := multimodalRedirectPayload("document")
			modify(payload)
			response := requestRedirect(t, redirectTestRouter(dispatcher), "multimodal_embedding/invoke", payload)
			assert.Equal(t, http.StatusBadRequest, response.Code)
			assert.NotContains(t, response.Body.String(), redirectTestSecret)
		})
	}
}

func TestMultimodalRedirectRechecksTargetPinAfterClusterReplay(t *testing.T) {
	entry := redirectTestDispatcher(t)
	target := redirectTestDispatcher(t)
	var targetLookups, targetCalls, entryCalls atomic.Int32
	target.installation = func(tenant, plugin string) (*models.PluginInstallation, error) {
		targetLookups.Add(1)
		return &models.PluginInstallation{TenantID: tenant, PluginID: plugin,
			PluginUniqueIdentifier: "example/target:2.0.0@bbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbb"}, nil
	}
	operation := target.operations["multimodal_embedding/invoke"]
	operation.invoke = func(*gin.Context, *modelRedirectRequest, []byte) { targetCalls.Add(1) }
	target.operations["multimodal_embedding/invoke"] = operation
	entryOperation := entry.operations["multimodal_embedding/invoke"]
	entryOperation.invoke = func(*gin.Context, *modelRedirectRequest, []byte) { entryCalls.Add(1) }
	entry.operations["multimodal_embedding/invoke"] = entryOperation
	bodies := make(chan []byte, 1)
	targetServer := multimodalObservedTarget(t, redirectTestRouter(target), bodies)
	entryServer := httptest.NewServer(redirectTestRouter(entry, multimodalTestRelay(t, targetServer.URL, nil)))
	t.Cleanup(entryServer.Close)
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	request, original := multimodalEntryRequest(t, ctx, entryServer.URL, "query")
	response, err := http.DefaultClient.Do(request)
	require.NoError(t, err)
	defer response.Body.Close()
	body, err := io.ReadAll(response.Body)
	require.NoError(t, err)
	assert.Equal(t, original, <-bodies)
	assert.Equal(t, http.StatusConflict, response.StatusCode)
	assert.Contains(t, string(body), "redirect_target_version_mismatch")
	assert.NotContains(t, string(body), redirectTestSecret)
	assert.Equal(t, int32(1), targetLookups.Load())
	assert.Zero(t, targetCalls.Load())
	assert.Zero(t, entryCalls.Load())
}

func TestMultimodalRedirectActualBodyRelayPreservesTypedDocumentAndQuery(t *testing.T) {
	for _, inputType := range []string{"document", "query"} {
		t.Run(inputType, func(t *testing.T) {
			entry := redirectTestDispatcher(t)
			target := redirectTestDispatcher(t)
			var targetCalls, entryCalls atomic.Int32
			target.operations["multimodal_embedding/invoke"] = redirectedOperation([]string{"text-embedding"}, 10,
				func(request *plugin_entities.InvokePluginRequest[requests.RequestInvokeMultimodalEmbedding], ctx *gin.Context, timeout int) {
					targetCalls.Add(1)
					assertMultimodalTarget(t, request, inputType)
					assert.Equal(t, 10, timeout)
					ctx.Header("X-Target-Contract", "multimodal")
					ctx.JSON(http.StatusOK, gin.H{"embeddings": [][]float64{{1, 2}, {3, 4}}})
				})
			operation := entry.operations["multimodal_embedding/invoke"]
			operation.invoke = func(*gin.Context, *modelRedirectRequest, []byte) { entryCalls.Add(1) }
			entry.operations["multimodal_embedding/invoke"] = operation
			bodies := make(chan []byte, 1)
			targetServer := multimodalObservedTarget(t, redirectTestRouter(target), bodies)
			entryServer := httptest.NewServer(redirectTestRouter(entry, multimodalTestRelay(t, targetServer.URL, nil)))
			t.Cleanup(entryServer.Close)
			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer cancel()
			request, original := multimodalEntryRequest(t, ctx, entryServer.URL, inputType)
			response, err := http.DefaultClient.Do(request)
			require.NoError(t, err)
			defer response.Body.Close()
			body, err := io.ReadAll(response.Body)
			require.NoError(t, err)
			assert.Equal(t, original, <-bodies)
			assert.Equal(t, http.StatusOK, response.StatusCode)
			assert.Equal(t, "multimodal", response.Header.Get("X-Target-Contract"))
			assert.JSONEq(t, `{"embeddings":[[1,2],[3,4]]}`, string(body))
			assert.NotContains(t, string(body), redirectTestSecret)
			assert.NotContains(t, fmt.Sprint(response.Header), redirectTestSecret)
			assert.Equal(t, int32(1), targetCalls.Load())
			assert.Zero(t, entryCalls.Load())
		})
	}
}

func TestMultimodalRedirectActualBodyRelayCancellationClosesTargetAndResponse(t *testing.T) {
	entry := redirectTestDispatcher(t)
	target := redirectTestDispatcher(t)
	var targetCalls, entryCalls atomic.Int32
	cancelled, closed, stop := make(chan struct{}), make(chan struct{}), make(chan struct{})
	target.operations["multimodal_embedding/invoke"] = redirectedOperation([]string{"text-embedding"}, 10,
		func(request *plugin_entities.InvokePluginRequest[requests.RequestInvokeMultimodalEmbedding], ctx *gin.Context, _ int) {
			targetCalls.Add(1)
			assertMultimodalTarget(t, request, "query")
			ctx.Header("Content-Type", "text/event-stream")
			ctx.Header("X-Target-Contract", "multimodal-cancel")
			ctx.String(http.StatusOK, "data: {\"code\":0,\"data\":{\"embeddings\":[[1,2]]}}\n\n")
			ctx.Writer.Flush()
			select {
			case <-ctx.Request.Context().Done():
				close(cancelled)
			case <-stop:
			}
		})
	operation := entry.operations["multimodal_embedding/invoke"]
	operation.invoke = func(*gin.Context, *modelRedirectRequest, []byte) { entryCalls.Add(1) }
	entry.operations["multimodal_embedding/invoke"] = operation
	bodies := make(chan []byte, 1)
	targetServer := multimodalObservedTarget(t, redirectTestRouter(target), bodies)
	entryServer := httptest.NewServer(redirectTestRouter(entry, multimodalTestRelay(t, targetServer.URL, func() { close(closed) })))
	t.Cleanup(entryServer.Close)
	t.Cleanup(func() { close(stop) })
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	request, original := multimodalEntryRequest(t, ctx, entryServer.URL, "query")
	response, err := http.DefaultClient.Do(request)
	require.NoError(t, err)
	defer response.Body.Close()
	assert.Equal(t, original, <-bodies)
	assert.Equal(t, http.StatusOK, response.StatusCode)
	assert.Equal(t, "text/event-stream", response.Header.Get("Content-Type"))
	assert.Equal(t, "multimodal-cancel", response.Header.Get("X-Target-Contract"))
	line, err := bufio.NewReader(response.Body).ReadString('\n')
	require.NoError(t, err)
	assert.Contains(t, line, `"embeddings":[[1,2]]`)
	assert.NotContains(t, line, redirectTestSecret)
	cancel()
	for name, event := range map[string]<-chan struct{}{"target cancellation": cancelled, "relay response close": closed} {
		select {
		case <-event:
		case <-time.After(3 * time.Second):
			t.Fatalf("%s did not finish", name)
		}
	}
	assert.Equal(t, int32(1), targetCalls.Load())
	assert.Zero(t, entryCalls.Load())
}

func TestMultimodalRedirectActualBodyRelayPreservesErrorsWithoutFallback(t *testing.T) {
	for _, test := range []struct {
		name, contentType, body string
		status                  int
	}{
		{"bad_request", "application/json", `{"code":"synthetic-invalid-documents","message":"synthetic-invalid-documents"}`, http.StatusBadRequest},
		{"unavailable", "application/json", `{"code":"synthetic-target-unavailable","message":"synthetic-target-unavailable"}`, http.StatusServiceUnavailable},
		{"sse_error", "text/event-stream", "data: {\"code\":-500,\"message\":\"synthetic-target-error\",\"data\":null}\n\n", http.StatusOK},
	} {
		t.Run(test.name, func(t *testing.T) {
			entry := redirectTestDispatcher(t)
			target := redirectTestDispatcher(t)
			var targetCalls, entryCalls atomic.Int32
			target.operations["multimodal_embedding/invoke"] = redirectedOperation([]string{"text-embedding"}, 10,
				func(request *plugin_entities.InvokePluginRequest[requests.RequestInvokeMultimodalEmbedding], ctx *gin.Context, _ int) {
					targetCalls.Add(1)
					assertMultimodalTarget(t, request, "document")
					ctx.Header("X-Target-Contract", "multimodal-error")
					ctx.Data(test.status, test.contentType, []byte(test.body))
				})
			operation := entry.operations["multimodal_embedding/invoke"]
			operation.invoke = func(*gin.Context, *modelRedirectRequest, []byte) { entryCalls.Add(1) }
			entry.operations["multimodal_embedding/invoke"] = operation
			bodies := make(chan []byte, 1)
			targetServer := multimodalObservedTarget(t, redirectTestRouter(target), bodies)
			entryServer := httptest.NewServer(redirectTestRouter(entry, multimodalTestRelay(t, targetServer.URL, nil)))
			t.Cleanup(entryServer.Close)
			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer cancel()
			request, original := multimodalEntryRequest(t, ctx, entryServer.URL, "document")
			response, err := http.DefaultClient.Do(request)
			require.NoError(t, err)
			defer response.Body.Close()
			body, err := io.ReadAll(response.Body)
			require.NoError(t, err)
			assert.Equal(t, original, <-bodies)
			assert.Equal(t, test.status, response.StatusCode)
			assert.Equal(t, test.contentType, response.Header.Get("Content-Type"))
			assert.Equal(t, "multimodal-error", response.Header.Get("X-Target-Contract"))
			assert.Equal(t, test.body, string(body))
			assert.NotContains(t, string(body), redirectTestSecret)
			assert.NotContains(t, fmt.Sprint(response.Header), redirectTestSecret)
			assert.Equal(t, int32(1), targetCalls.Load())
			assert.Zero(t, entryCalls.Load())
		})
	}
}

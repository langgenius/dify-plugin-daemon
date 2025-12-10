package integration_test

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"net/http"
	"os"
	"strings"
	"testing"
	"time"

	e2eutils "github.com/langgenius/dify-plugin-daemon/integration/e2e/testutils"
	"github.com/langgenius/dify-plugin-daemon/internal/core/io_tunnel/access_types"
	coretestutils "github.com/langgenius/dify-plugin-daemon/internal/core/testutils"
	"github.com/langgenius/dify-plugin-daemon/pkg/entities/model_entities"
	"github.com/langgenius/dify-plugin-daemon/pkg/entities/requests"
	"github.com/langgenius/dify-plugin-daemon/pkg/utils/routine"
	"github.com/langgenius/dify-plugin-daemon/pkg/utils/stream"
	"github.com/stretchr/testify/require"
)

const (
	defaultSourceRepo   = "langgenius/dify-official-plugins"
	defaultSourceSubdir = "models/ollama"
	defaultSourceRef    = "main"
	runtimeWorkdir      = "./integration_e2e_runtime"
)

func TestOllamaPluginPackageAndInstall(t *testing.T) {
	routine.InitPool(10000)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()

	owner, repo := parseRepo(t, getenv("OLLAMA_SOURCE_REPO", defaultSourceRepo))
	ref := getenv("OLLAMA_SOURCE_REF", defaultSourceRef)
	subdir := strings.Trim(getenv("OLLAMA_SOURCE_SUBDIR", defaultSourceSubdir), "/")

	archive, resolvedRef, err := e2eutils.DownloadRepoArchive(ctx, owner, repo, ref)
	require.NoError(t, err, "failed to download repo archive")
	t.Logf("downloaded source %s/%s@%s", owner, repo, resolvedRef)

	pluginRoot, cleanup, err := e2eutils.ExtractPluginSource(archive, subdir)
	require.NoError(t, err, "failed to extract plugin source")
	defer cleanup()

	packageBytes, manifest, err := e2eutils.PackPluginFromDir(pluginRoot)
	require.NoError(t, err, "failed to pack plugin")
	require.NotEmpty(t, packageBytes, "packed plugin is empty")
	t.Logf("packed plugin %s", manifest.Identity())

	port, shutdown := startFakeLLMServer()
	defer shutdown()
	baseURL := fmt.Sprintf("http://127.0.0.1:%d", port)
	t.Logf("started fake LLM server at %s", baseURL)

	require.NoError(t, os.MkdirAll(runtimeWorkdir, 0o755))
	defer coretestutils.ClearTestingPath(runtimeWorkdir)

	runtime, err := coretestutils.GetRuntime(packageBytes, runtimeWorkdir, 1)
	require.NoError(t, err, "failed to start plugin runtime")
	defer runtime.GracefulStop(false)
	t.Log("plugin runtime ready")

	credentials := requests.Credentials{
		Credentials: map[string]any{
			"base_url":              baseURL,
			"api_key":               "test-key",
			"mode":                  "chat",
			"context_size":          "4096",
			"max_tokens":            "4096",
			"vision_support":        "false",
			"function_call_support": "false",
		},
	}

	validateStream, err := coretestutils.RunOnce[requests.RequestValidateProviderCredentials, map[string]any](
		runtime,
		access_types.PLUGIN_ACCESS_TYPE_MODEL,
		access_types.PLUGIN_ACCESS_ACTION_VALIDATE_PROVIDER_CREDENTIALS,
		requests.RequestValidateProviderCredentials{
			Credentials: credentials,
			Provider:    "ollama",
		},
	)
	require.NoError(t, err, "failed to validate credentials")
	drainStream(t, validateStream)
	t.Log("credentials validated successfully")

	llmStream, err := coretestutils.RunOnce[requests.RequestInvokeLLM, model_entities.LLMResultChunk](
		runtime,
		access_types.PLUGIN_ACCESS_TYPE_MODEL,
		access_types.PLUGIN_ACCESS_ACTION_INVOKE_LLM,
		requests.RequestInvokeLLM{
			BaseRequestInvokeModel: requests.BaseRequestInvokeModel{
				Provider: "ollama",
				Model:    getenv("OLLAMA_MODEL", "llama3"),
			},
			Credentials: credentials,
			InvokeLLMSchema: requests.InvokeLLMSchema{
				ModelParameters: map[string]any{},
				PromptMessages: []model_entities.PromptMessage{
					{
						Role:    model_entities.PROMPT_MESSAGE_ROLE_USER,
						Content: "Hello from E2E packaging test",
					},
				},
				Stream: true,
			},
			ModelType: model_entities.MODEL_TYPE_LLM,
		},
	)
	require.NoError(t, err, "failed to invoke LLM")

	sawToken := false
	for llmStream.Next() {
		chunk, readErr := llmStream.Read()
		require.NoError(t, readErr, "failed to read stream response")

		if chunk.Delta.Message.Content != nil {
			switch content := chunk.Delta.Message.Content.(type) {
			case string:
				if strings.TrimSpace(content) != "" {
					sawToken = true
				}
			case []model_entities.PromptMessageContent:
				if len(content) > 0 {
					sawToken = true
				}
			}
		}
	}

	t.Log("LLM invocation completed successfully")
	require.True(t, sawToken, "no LLM output received")
}

func parseRepo(t *testing.T, repo string) (string, string) {
	t.Helper()
	parts := strings.Split(strings.TrimSpace(repo), "/")
	require.Len(t, parts, 2, "repo format should be owner/repo")
	return parts[0], parts[1]
}

func getenv(key, fallback string) string {
	if v := strings.TrimSpace(os.Getenv(key)); v != "" {
		return v
	}
	return fallback
}

func drainStream[T any](t *testing.T, s *stream.Stream[T]) {
	t.Helper()
	for s.Next() {
		_, err := s.Read()
		require.NoError(t, err)
	}
}

// startFakeLLMServer starts a mock server supporting both Ollama and OpenAI API formats.
func startFakeLLMServer() (int, func()) {
	listener, err := net.Listen("tcp", ":0")
	if err != nil {
		panic(fmt.Sprintf("failed to find available port: %v", err))
	}
	port := listener.Addr().(*net.TCPAddr).Port
	listener.Close()

	mux := http.NewServeMux()
	server := &http.Server{
		Addr:    fmt.Sprintf(":%d", port),
		Handler: mux,
	}

	// Ollama API: /api/chat (streaming)
	mux.HandleFunc("/api/chat", func(w http.ResponseWriter, r *http.Request) {
		body, _ := io.ReadAll(r.Body)
		fmt.Printf("[Mock] POST /api/chat\n")
		fmt.Printf("[Mock] Request: %s\n", string(body))

		w.Header().Set("Content-Type", "application/x-ndjson")
		w.Header().Set("Transfer-Encoding", "chunked")

		flusher, ok := w.(http.Flusher)
		if !ok {
			http.Error(w, "streaming not supported", http.StatusInternalServerError)
			return
		}

		words := []string{"Hello", " from", " fake", " Ollama", " server", "!"}
		for i, word := range words {
			resp := map[string]any{
				"model": "llama3",
				"message": map[string]any{
					"role":    "assistant",
					"content": word,
				},
				"done": i == len(words)-1,
			}
			if i == len(words)-1 {
				resp["prompt_eval_count"] = 10
				resp["eval_count"] = len(words)
			}
			data, _ := json.Marshal(resp)
			fmt.Printf("[Mock] Response chunk: %s\n", string(data))
			fmt.Fprintf(w, "%s\n", data)
			flusher.Flush()
			time.Sleep(10 * time.Millisecond)
		}
		fmt.Printf("[Mock] /api/chat completed\n")
	})

	// Ollama API: /api/generate (streaming)
	mux.HandleFunc("/api/generate", func(w http.ResponseWriter, r *http.Request) {
		body, _ := io.ReadAll(r.Body)
		fmt.Printf("[Mock] POST /api/generate\n")
		fmt.Printf("[Mock] Request: %s\n", string(body))

		w.Header().Set("Content-Type", "application/x-ndjson")
		flusher, _ := w.(http.Flusher)

		words := []string{"Hello", " from", " generate", "!"}
		for i, word := range words {
			resp := map[string]any{
				"model":    "llama3",
				"response": word,
				"done":     i == len(words)-1,
			}
			data, _ := json.Marshal(resp)
			fmt.Printf("[Mock] Response chunk: %s\n", string(data))
			fmt.Fprintf(w, "%s\n", data)
			if flusher != nil {
				flusher.Flush()
			}
			time.Sleep(10 * time.Millisecond)
		}
		fmt.Printf("[Mock] /api/generate completed\n")
	})

	// OpenAI API: /v1/chat/completions (streaming)
	mux.HandleFunc("/v1/chat/completions", func(w http.ResponseWriter, r *http.Request) {
		body, _ := io.ReadAll(r.Body)
		fmt.Printf("[Mock] POST /v1/chat/completions\n")
		fmt.Printf("[Mock] Request: %s\n", string(body))

		w.Header().Set("Content-Type", "text/event-stream")
		w.Header().Set("Cache-Control", "no-cache")
		flusher, _ := w.(http.Flusher)

		words := []string{"Hello", " from", " OpenAI", " compatible", "!"}
		for i, word := range words {
			var finishReason *string
			if i == len(words)-1 {
				s := "stop"
				finishReason = &s
			}
			resp := map[string]any{
				"id":      "chatcmpl-test",
				"object":  "chat.completion.chunk",
				"created": time.Now().Unix(),
				"model":   "gpt-3.5-turbo",
				"choices": []map[string]any{
					{
						"index":         0,
						"delta":         map[string]any{"content": word},
						"finish_reason": finishReason,
					},
				},
			}
			data, _ := json.Marshal(resp)
			fmt.Printf("[Mock] Response chunk: %s\n", string(data))
			fmt.Fprintf(w, "data: %s\n\n", data)
			if flusher != nil {
				flusher.Flush()
			}
			time.Sleep(10 * time.Millisecond)
		}
		fmt.Fprintf(w, "data: [DONE]\n\n")
		if flusher != nil {
			flusher.Flush()
		}
		fmt.Printf("[Mock] /v1/chat/completions completed\n")
	})

	go func() {
		if err := server.ListenAndServe(); err != nil && !strings.Contains(err.Error(), "Server closed") {
			fmt.Printf("[Mock] server error: %v\n", err)
		}
	}()

	return port, func() {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		server.Shutdown(ctx)
	}
}

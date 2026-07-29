package io_tunnel

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/langgenius/dify-plugin-daemon/internal/core/io_tunnel/access_types"
	"github.com/langgenius/dify-plugin-daemon/internal/core/serverless_runtime"
	"github.com/langgenius/dify-plugin-daemon/internal/core/session_manager"
	"github.com/langgenius/dify-plugin-daemon/pkg/entities"
	"github.com/langgenius/dify-plugin-daemon/pkg/entities/model_entities"
	"github.com/langgenius/dify-plugin-daemon/pkg/entities/plugin_entities"
	"github.com/langgenius/dify-plugin-daemon/pkg/entities/requests"
	"github.com/langgenius/dify-plugin-daemon/pkg/utils/parser"
	"github.com/langgenius/dify-plugin-daemon/pkg/utils/routine"
	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/require"
	"go.opentelemetry.io/otel/sdk/metric/metricdata"
)

type tokenBatchRuntime struct {
	mu              sync.Mutex
	listener        *entities.Broadcast[plugin_entities.SessionMessage]
	runtimeType     plugin_entities.PluginRuntimeType
	maxRequestBytes int
	payloads        [][]byte
	inFlight        int
	maxInFlight     int
	responseFn      func(int, []string) []plugin_entities.SessionMessage
}

func (r *tokenBatchRuntime) Type() plugin_entities.PluginRuntimeType {
	if r.runtimeType != "" {
		return r.runtimeType
	}
	return plugin_entities.PLUGIN_RUNTIME_TYPE_SERVERLESS
}

func (r *tokenBatchRuntime) Configuration() *plugin_entities.PluginDeclaration {
	return &plugin_entities.PluginDeclaration{}
}

func (r *tokenBatchRuntime) Identity() (plugin_entities.PluginUniqueIdentifier, error) {
	return testPluginUniqueIdentifier, nil
}

func (r *tokenBatchRuntime) HashedIdentity() (string, error) {
	return plugin_entities.HashedIdentity(testPluginUniqueIdentifier.String()), nil
}

func (r *tokenBatchRuntime) Checksum() (string, error) {
	return testPluginUniqueIdentifier.Checksum(), nil
}

func (r *tokenBatchRuntime) MaxServerlessRequestBytes() int {
	return r.maxRequestBytes
}

func (r *tokenBatchRuntime) Listen(string) (*entities.Broadcast[plugin_entities.SessionMessage], error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	r.listener = entities.NewCallbackHandler[plugin_entities.SessionMessage]()
	return r.listener, nil
}

func (r *tokenBatchRuntime) Write(
	_ string,
	_ access_types.PluginAccessAction,
	data []byte,
) error {
	r.mu.Lock()
	listener := r.listener
	r.payloads = append(r.payloads, append([]byte(nil), data...))
	callIndex := len(r.payloads) - 1
	r.inFlight++
	if r.inFlight > r.maxInFlight {
		r.maxInFlight = r.inFlight
	}
	r.mu.Unlock()

	var message struct {
		Data struct {
			Texts []string `json:"texts"`
		} `json:"data"`
	}
	if err := json.Unmarshal(data, &message); err != nil {
		return err
	}

	go func() {
		messages := r.defaultResponseMessages(message.Data.Texts)
		if r.responseFn != nil {
			messages = r.responseFn(callIndex, message.Data.Texts)
		}

		for _, message := range messages {
			if message.Type == plugin_entities.SESSION_MESSAGE_TYPE_END ||
				message.Type == plugin_entities.SESSION_MESSAGE_TYPE_ERROR {
				r.mu.Lock()
				r.inFlight--
				r.mu.Unlock()
			}
			listener.Send(message)
		}
	}()

	return nil
}

func (r *tokenBatchRuntime) defaultResponseMessages(texts []string) []plugin_entities.SessionMessage {
	numTokens := make([]int, len(texts))
	for i, text := range texts {
		numTokens[i] = len([]rune(text))
	}
	return []plugin_entities.SessionMessage{
		{
			Type: plugin_entities.SESSION_MESSAGE_TYPE_STREAM,
			Data: parser.MarshalJsonBytes(model_entities.GetTextEmbeddingNumTokensResponse{
				NumTokens: numTokens,
			}),
		},
		{
			Type: plugin_entities.SESSION_MESSAGE_TYPE_END,
		},
	}
}

type embeddingBatchRuntime struct {
	tokenBatchRuntime
	embeddingResponseFn func(int, []string) []plugin_entities.SessionMessage
}

func (r *embeddingBatchRuntime) Write(
	_ string,
	_ access_types.PluginAccessAction,
	data []byte,
) error {
	r.mu.Lock()
	listener := r.listener
	r.payloads = append(r.payloads, append([]byte(nil), data...))
	callIndex := len(r.payloads) - 1
	r.inFlight++
	if r.inFlight > r.maxInFlight {
		r.maxInFlight = r.inFlight
	}
	r.mu.Unlock()

	var message struct {
		Data struct {
			Texts []string `json:"texts"`
		} `json:"data"`
	}
	if err := json.Unmarshal(data, &message); err != nil {
		return err
	}

	go func() {
		messages := embeddingResponseMessages(message.Data.Texts)
		if r.embeddingResponseFn != nil {
			messages = r.embeddingResponseFn(callIndex, message.Data.Texts)
		}
		for _, message := range messages {
			if message.Type == plugin_entities.SESSION_MESSAGE_TYPE_END ||
				message.Type == plugin_entities.SESSION_MESSAGE_TYPE_ERROR {
				r.mu.Lock()
				r.inFlight--
				r.mu.Unlock()
			}
			listener.Send(message)
		}
	}()
	return nil
}

func embeddingResponseMessages(texts []string) []plugin_entities.SessionMessage {
	return []plugin_entities.SessionMessage{
		{
			Type: plugin_entities.SESSION_MESSAGE_TYPE_STREAM,
			Data: parser.MarshalJsonBytes(embeddingResult(texts)),
		},
		{Type: plugin_entities.SESSION_MESSAGE_TYPE_END},
	}
}

func embeddingResult(texts []string) model_entities.TextEmbeddingResult {
	tokens := len(texts)
	totalTokens := len(texts) * 10
	currency := "USD"
	latency := float64(len(texts)) * 0.5
	embeddings := make([][]float64, len(texts))
	for i, text := range texts {
		embeddings[i] = []float64{float64(len([]rune(text)))}
	}
	return model_entities.TextEmbeddingResult{
		Model:      "embedding-model",
		Embeddings: embeddings,
		Usage: model_entities.EmbeddingUsage{
			Tokens:      &tokens,
			TotalTokens: &totalTokens,
			UnitPrice:   decimal.RequireFromString("0.01"),
			PriceUnit:   decimal.NewFromInt(1000),
			TotalPrice:  decimal.RequireFromString("0.02").Mul(decimal.NewFromInt(int64(len(texts)))),
			Currency:    &currency,
			Latency:     &latency,
		},
	}
}

func TestGetTextEmbeddingNumTokensServerlessAwareSplitsByFinalPayloadBytes(t *testing.T) {
	request := newTokenCountRequest([]string{`plain`, `中文`, `emoji 😀`, `quote " slash \`})
	runtime := &tokenBatchRuntime{}
	session := newTestSession(
		t,
		runtime,
		access_types.PLUGIN_ACCESS_TYPE_MODEL,
		access_types.PLUGIN_ACCESS_ACTION_GET_TEXT_EMBEDDING_NUM_TOKENS,
	)

	twoItemsRequest := *request
	twoItemsRequest.Texts = request.Texts[:2]
	runtime.maxRequestBytes = maxTokenCountSingleItemPayloadSize(session, request)
	runtime.maxRequestBytes = max(runtime.maxRequestBytes, tokenCountPayloadSize(session, &twoItemsRequest))

	response, err := getTextEmbeddingNumTokensServerlessAware(session, request)
	require.NoError(t, err)

	results := readAllStreamValues(t, response)
	require.Equal(t, []model_entities.GetTextEmbeddingNumTokensResponse{{
		NumTokens: []int{5, 2, 7, 15},
	}}, results)

	runtime.mu.Lock()
	defer runtime.mu.Unlock()
	require.Len(t, runtime.payloads, 3)
	for _, payload := range runtime.payloads {
		require.LessOrEqual(t, len(payload), runtime.maxRequestBytes)
	}
	require.Equal(t, 1, runtime.maxInFlight)
}

func TestGetTextEmbeddingNumTokensServerlessAwareUsesRealRuntimeSerially(t *testing.T) {
	routine.InitPool(4)
	var (
		mu          sync.Mutex
		inFlight    int
		maxInFlight int
		received    []string
		handlerErr  error
	)
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var message struct {
			Data struct {
				Texts []string `json:"texts"`
			} `json:"data"`
		}
		if err := json.NewDecoder(r.Body).Decode(&message); err != nil {
			mu.Lock()
			handlerErr = err
			mu.Unlock()
			return
		}

		mu.Lock()
		inFlight++
		maxInFlight = max(maxInFlight, inFlight)
		received = append(received, message.Data.Texts...)
		mu.Unlock()
		defer func() {
			mu.Lock()
			inFlight--
			mu.Unlock()
		}()

		numTokens := make([]int, len(message.Data.Texts))
		for i, text := range message.Data.Texts {
			numTokens[i] = len([]rune(text))
		}
		encoder := json.NewEncoder(w)
		if err := encoder.Encode(map[string]any{
			"session_id": r.Header.Get("Dify-Plugin-Session-ID"),
			"event":      "session",
			"data": plugin_entities.SessionMessage{
				Type: plugin_entities.SESSION_MESSAGE_TYPE_STREAM,
				Data: parser.MarshalJsonBytes(model_entities.GetTextEmbeddingNumTokensResponse{
					NumTokens: numTokens,
				}),
			},
		}); err != nil {
			mu.Lock()
			handlerErr = err
			mu.Unlock()
			return
		}
		if err := encoder.Encode(map[string]any{
			"session_id": r.Header.Get("Dify-Plugin-Session-ID"),
			"event":      "session",
			"data": plugin_entities.SessionMessage{
				Type: plugin_entities.SESSION_MESSAGE_TYPE_END,
			},
		}); err != nil {
			mu.Lock()
			handlerErr = err
			mu.Unlock()
		}
	}))
	t.Cleanup(server.Close)

	runtime := &serverless_runtime.ServerlessPluginRuntime{
		Client:                    server.Client(),
		LambdaURL:                 server.URL,
		MaxRetryTimes:             1,
		PluginMaxExecutionTimeout: 10,
		RuntimeBufferSize:         1024,
		RuntimeMaxBufferSize:      1024,
	}
	request := newTokenCountRequest([]string{"one", "two", "three"})
	session := newTestSession(
		t,
		runtime,
		access_types.PLUGIN_ACCESS_TYPE_MODEL,
		access_types.PLUGIN_ACCESS_ACTION_GET_TEXT_EMBEDDING_NUM_TOKENS,
	)
	runtime.MaxRequestBytes = maxTokenCountSingleItemPayloadSize(session, request)

	response, err := getTextEmbeddingNumTokensServerlessAware(session, request)
	require.NoError(t, err)
	results := readAllStreamValues(t, response)

	require.Equal(t, []model_entities.GetTextEmbeddingNumTokensResponse{{
		NumTokens: []int{3, 3, 5},
	}}, results)
	mu.Lock()
	defer mu.Unlock()
	require.NoError(t, handlerErr)
	require.Equal(t, []string{"one", "two", "three"}, received)
	require.Equal(t, 1, maxInFlight)
}

func TestGetTextEmbeddingNumTokensServerlessAwareCancelsRealRuntime(t *testing.T) {
	routine.InitPool(2)
	requestStarted := make(chan struct{}, 1)
	requestCanceled := make(chan struct{}, 1)
	releaseHandler := make(chan struct{})
	var (
		mu           sync.Mutex
		requestCount int
	)
	server := httptest.NewServer(http.HandlerFunc(func(_ http.ResponseWriter, r *http.Request) {
		_, _ = io.Copy(io.Discard, r.Body)
		mu.Lock()
		requestCount++
		mu.Unlock()
		select {
		case requestStarted <- struct{}{}:
		default:
		}
		select {
		case <-r.Context().Done():
			select {
			case requestCanceled <- struct{}{}:
			default:
			}
		case <-releaseHandler:
		}
	}))
	t.Cleanup(func() {
		close(releaseHandler)
		server.Close()
	})

	runtime := &serverless_runtime.ServerlessPluginRuntime{
		Client:                    server.Client(),
		LambdaURL:                 server.URL,
		MaxRetryTimes:             1,
		PluginMaxExecutionTimeout: 10,
		RuntimeBufferSize:         1024,
		RuntimeMaxBufferSize:      1024,
	}
	request := newTokenCountRequest([]string{"one", "two"})
	session := newTestSession(
		t,
		runtime,
		access_types.PLUGIN_ACCESS_TYPE_MODEL,
		access_types.PLUGIN_ACCESS_ACTION_GET_TEXT_EMBEDDING_NUM_TOKENS,
	)
	runtime.MaxRequestBytes = maxTokenCountSingleItemPayloadSize(session, request)

	response, err := getTextEmbeddingNumTokensServerlessAware(session, request)
	require.NoError(t, err)
	<-requestStarted
	response.Close()

	select {
	case <-requestCanceled:
	case <-time.After(time.Second):
		t.Fatal("server request did not observe outer stream cancellation")
	}
	mu.Lock()
	defer mu.Unlock()
	require.Equal(t, 1, requestCount)
}

func TestGetTextEmbeddingNumTokensServerlessAwareDoesNotMutateRequest(t *testing.T) {
	originalTexts := []string{"one", "two", "three"}
	request := newTokenCountRequest(originalTexts)
	runtime := &tokenBatchRuntime{}
	session := newTestSession(
		t,
		runtime,
		access_types.PLUGIN_ACCESS_TYPE_MODEL,
		access_types.PLUGIN_ACCESS_ACTION_GET_TEXT_EMBEDDING_NUM_TOKENS,
	)

	runtime.maxRequestBytes = maxTokenCountSingleItemPayloadSize(session, request)

	response, err := getTextEmbeddingNumTokensServerlessAware(session, request)
	require.NoError(t, err)
	readAllStreamValues(t, response)

	require.Equal(t, originalTexts, request.Texts)
}

func TestGetTextEmbeddingNumTokensServerlessAwareRejectsOversizeSingleItem(t *testing.T) {
	request := newTokenCountRequest([]string{"too large"})
	runtime := &tokenBatchRuntime{}
	session := newTestSession(
		t,
		runtime,
		access_types.PLUGIN_ACCESS_TYPE_MODEL,
		access_types.PLUGIN_ACCESS_ACTION_GET_TEXT_EMBEDDING_NUM_TOKENS,
	)
	runtime.maxRequestBytes = tokenCountPayloadSize(session, request) - 1

	response, err := getTextEmbeddingNumTokensServerlessAware(session, request)

	require.Nil(t, response)
	require.ErrorContains(t, err, "ServerlessPayloadTooLarge")
	runtime.mu.Lock()
	defer runtime.mu.Unlock()
	require.Empty(t, runtime.payloads)
}

func TestGetTextEmbeddingNumTokensServerlessAwareRejectsOversizeFixedEnvelope(t *testing.T) {
	request := newTokenCountRequest([]string{"small"})
	request.Credentials.Credentials["api_key"] = strings.Repeat("x", 1024)
	runtime := &tokenBatchRuntime{}
	session := newTestSession(
		t,
		runtime,
		access_types.PLUGIN_ACCESS_TYPE_MODEL,
		access_types.PLUGIN_ACCESS_ACTION_GET_TEXT_EMBEDDING_NUM_TOKENS,
	)
	emptyRequest := *request
	emptyRequest.Texts = []string{}
	runtime.maxRequestBytes = tokenCountPayloadSize(session, &emptyRequest) - 1

	response, err := getTextEmbeddingNumTokensServerlessAware(session, request)

	require.Nil(t, response)
	var payloadTooLarge *serverless_runtime.ServerlessPayloadTooLargeError
	require.ErrorAs(t, err, &payloadTooLarge)
	require.Nil(t, payloadTooLarge.ItemIndex)
	runtime.mu.Lock()
	defer runtime.mu.Unlock()
	require.Empty(t, runtime.payloads)
}

func TestGetTextEmbeddingNumTokensServerlessAwareStopsAfterBatchError(t *testing.T) {
	request := newTokenCountRequest([]string{"one", "two", "three"})
	runtime := &tokenBatchRuntime{
		responseFn: func(callIndex int, texts []string) []plugin_entities.SessionMessage {
			if callIndex == 1 {
				return []plugin_entities.SessionMessage{{
					Type: plugin_entities.SESSION_MESSAGE_TYPE_ERROR,
					Data: parser.MarshalJsonBytes(plugin_entities.ErrorResponse{
						ErrorType: "ProviderError",
						Message:   "second batch failed",
					}),
				}}
			}
			return (&tokenBatchRuntime{}).defaultResponseMessages(texts)
		},
	}
	session := newTestSession(
		t,
		runtime,
		access_types.PLUGIN_ACCESS_TYPE_MODEL,
		access_types.PLUGIN_ACCESS_ACTION_GET_TEXT_EMBEDDING_NUM_TOKENS,
	)
	runtime.maxRequestBytes = maxTokenCountSingleItemPayloadSize(session, request)

	response, err := getTextEmbeddingNumTokensServerlessAware(session, request)
	require.NoError(t, err)

	values, streamErr := collectTokenCountStream(response)
	require.ErrorContains(t, streamErr, "second batch failed")
	require.Empty(t, values)

	runtime.mu.Lock()
	defer runtime.mu.Unlock()
	require.Len(t, runtime.payloads, 2)
}

func TestGetTextEmbeddingNumTokensServerlessAwareRejectsMismatchedBatchResult(t *testing.T) {
	request := newTokenCountRequest([]string{"one", "two", "three"})
	runtime := &tokenBatchRuntime{
		responseFn: func(callIndex int, texts []string) []plugin_entities.SessionMessage {
			if callIndex == 1 {
				return []plugin_entities.SessionMessage{
					{
						Type: plugin_entities.SESSION_MESSAGE_TYPE_STREAM,
						Data: parser.MarshalJsonBytes(model_entities.GetTextEmbeddingNumTokensResponse{
							NumTokens: []int{1, 2},
						}),
					},
					{Type: plugin_entities.SESSION_MESSAGE_TYPE_END},
				}
			}
			return (&tokenBatchRuntime{}).defaultResponseMessages(texts)
		},
	}
	session := newTestSession(
		t,
		runtime,
		access_types.PLUGIN_ACCESS_TYPE_MODEL,
		access_types.PLUGIN_ACCESS_ACTION_GET_TEXT_EMBEDDING_NUM_TOKENS,
	)
	runtime.maxRequestBytes = maxTokenCountSingleItemPayloadSize(session, request)

	response, err := getTextEmbeddingNumTokensServerlessAware(session, request)
	require.NoError(t, err)

	values, streamErr := collectTokenCountStream(response)
	require.ErrorContains(t, streamErr, "ServerlessBatchResultMismatch")
	require.Empty(t, values)

	runtime.mu.Lock()
	defer runtime.mu.Unlock()
	require.Len(t, runtime.payloads, 2)
}

func TestGetTextEmbeddingNumTokensServerlessAwareKeepsLocalRuntimeAsSingleCall(t *testing.T) {
	request := newTokenCountRequest([]string{"one", "two", "three"})
	runtime := &tokenBatchRuntime{
		runtimeType:     plugin_entities.PLUGIN_RUNTIME_TYPE_LOCAL,
		maxRequestBytes: 1,
	}
	session := newTestSession(
		t,
		runtime,
		access_types.PLUGIN_ACCESS_TYPE_MODEL,
		access_types.PLUGIN_ACCESS_ACTION_GET_TEXT_EMBEDDING_NUM_TOKENS,
	)

	response, err := getTextEmbeddingNumTokensServerlessAware(session, request)
	require.NoError(t, err)
	values := readAllStreamValues(t, response)

	require.Len(t, values, 1)
	runtime.mu.Lock()
	defer runtime.mu.Unlock()
	require.Len(t, runtime.payloads, 1)
}

func TestGetTextEmbeddingNumTokensServerlessAwareHonorsExactPayloadLimit(t *testing.T) {
	tests := []struct {
		name          string
		limitOffset   int
		expectedCalls int
	}{
		{
			name:          "exact limit remains one call",
			limitOffset:   0,
			expectedCalls: 1,
		},
		{
			name:          "one byte over the limit is split",
			limitOffset:   -1,
			expectedCalls: 2,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			request := newTokenCountRequest([]string{"one", "two"})
			runtime := &tokenBatchRuntime{}
			session := newTestSession(
				t,
				runtime,
				access_types.PLUGIN_ACCESS_TYPE_MODEL,
				access_types.PLUGIN_ACCESS_ACTION_GET_TEXT_EMBEDDING_NUM_TOKENS,
			)
			runtime.maxRequestBytes = tokenCountPayloadSize(session, request) + tt.limitOffset

			response, err := getTextEmbeddingNumTokensServerlessAware(session, request)
			require.NoError(t, err)
			readAllStreamValues(t, response)

			runtime.mu.Lock()
			defer runtime.mu.Unlock()
			require.Len(t, runtime.payloads, tt.expectedCalls)
		})
	}
}

func TestGetTextEmbeddingNumTokensServerlessAwareRequiresExactlyOneResponsePerBatch(t *testing.T) {
	tests := []struct {
		name      string
		responses int
	}{
		{name: "zero responses", responses: 0},
		{name: "multiple responses", responses: 2},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			request := newTokenCountRequest([]string{"one", "two"})
			runtime := &tokenBatchRuntime{
				responseFn: func(_ int, texts []string) []plugin_entities.SessionMessage {
					messages := make([]plugin_entities.SessionMessage, 0, tt.responses+1)
					for range tt.responses {
						messages = append(messages, plugin_entities.SessionMessage{
							Type: plugin_entities.SESSION_MESSAGE_TYPE_STREAM,
							Data: parser.MarshalJsonBytes(model_entities.GetTextEmbeddingNumTokensResponse{
								NumTokens: []int{len([]rune(texts[0]))},
							}),
						})
					}
					return append(messages, plugin_entities.SessionMessage{
						Type: plugin_entities.SESSION_MESSAGE_TYPE_END,
					})
				},
			}
			session := newTestSession(
				t,
				runtime,
				access_types.PLUGIN_ACCESS_TYPE_MODEL,
				access_types.PLUGIN_ACCESS_ACTION_GET_TEXT_EMBEDDING_NUM_TOKENS,
			)
			runtime.maxRequestBytes = maxTokenCountSingleItemPayloadSize(session, request)

			response, err := getTextEmbeddingNumTokensServerlessAware(session, request)
			require.NoError(t, err)

			values, streamErr := collectTokenCountStream(response)
			require.ErrorContains(t, streamErr, "ServerlessBatchResultMismatch")
			require.Empty(t, values)
		})
	}
}

func TestGetTextEmbeddingNumTokensServerlessAwareStopsAfterOuterCancellation(t *testing.T) {
	request := newTokenCountRequest([]string{"one", "two", "three"})
	firstBatchStarted := make(chan struct{}, 1)
	releaseFirstBatch := make(chan struct{})
	runtime := &tokenBatchRuntime{
		responseFn: func(callIndex int, texts []string) []plugin_entities.SessionMessage {
			if callIndex == 0 {
				firstBatchStarted <- struct{}{}
				<-releaseFirstBatch
			}
			return (&tokenBatchRuntime{}).defaultResponseMessages(texts)
		},
	}
	session := newTestSession(
		t,
		runtime,
		access_types.PLUGIN_ACCESS_TYPE_MODEL,
		access_types.PLUGIN_ACCESS_ACTION_GET_TEXT_EMBEDDING_NUM_TOKENS,
	)
	runtime.maxRequestBytes = maxTokenCountSingleItemPayloadSize(session, request)

	response, err := getTextEmbeddingNumTokensServerlessAware(session, request)
	require.NoError(t, err)
	<-firstBatchStarted

	response.Close()
	close(releaseFirstBatch)

	require.Eventually(t, func() bool {
		runtime.mu.Lock()
		defer runtime.mu.Unlock()
		return runtime.inFlight == 0
	}, time.Second, 10*time.Millisecond)
	runtime.mu.Lock()
	defer runtime.mu.Unlock()
	require.Len(t, runtime.payloads, 1)
}

func TestGetTextEmbeddingNumTokensServerlessAwareRecordsOneLogicalInvocation(t *testing.T) {
	reader := setupPluginInvocationMetricTest(t)
	request := newTokenCountRequest([]string{"one", "two", "three"})
	runtime := &tokenBatchRuntime{}
	session := newTestSession(
		t,
		runtime,
		access_types.PLUGIN_ACCESS_TYPE_MODEL,
		access_types.PLUGIN_ACCESS_ACTION_GET_TEXT_EMBEDDING_NUM_TOKENS,
	)
	runtime.maxRequestBytes = maxTokenCountSingleItemPayloadSize(session, request)

	response, err := getTextEmbeddingNumTokensServerlessAware(session, request)
	require.NoError(t, err)
	readAllStreamValues(t, response)

	metrics := collectPluginInvocationMetrics(t, reader)
	expectedAttrs := expectedPluginMetricAttributes(
		pluginInvocationOutcomeSuccess,
		string(plugin_entities.PLUGIN_RUNTIME_TYPE_SERVERLESS),
		string(access_types.PLUGIN_ACCESS_TYPE_MODEL),
	)
	counter, ok := pluginInvocationCounterValue(metrics, expectedAttrs)
	require.True(t, ok)
	require.EqualValues(t, 1, counter)

	batchCount, batchCountSum, ok := int64HistogramSummary(metrics, serverlessBatchCountMetricName)
	require.True(t, ok)
	require.EqualValues(t, 1, batchCount)
	require.EqualValues(t, 3, batchCountSum)

	payloadCount, _, ok := int64HistogramSummary(metrics, serverlessBatchPayloadMetricName)
	require.True(t, ok)
	require.EqualValues(t, 3, payloadCount)
}

func TestServerlessBatchLogsExcludeTextAndCredentials(t *testing.T) {
	var logOutput bytes.Buffer
	previousLogger := slog.Default()
	slog.SetDefault(slog.New(slog.NewJSONHandler(&logOutput, nil)))
	t.Cleanup(func() {
		slog.SetDefault(previousLogger)
	})

	const (
		sensitiveText       = "private-text-do-not-log"
		sensitiveCredential = "private-api-key-do-not-log"
	)
	request := newTokenCountRequest([]string{sensitiveText, "second"})
	request.Credentials.Credentials["api_key"] = sensitiveCredential
	runtime := &tokenBatchRuntime{}
	session := newTestSession(
		t,
		runtime,
		access_types.PLUGIN_ACCESS_TYPE_MODEL,
		access_types.PLUGIN_ACCESS_ACTION_GET_TEXT_EMBEDDING_NUM_TOKENS,
	)
	runtime.maxRequestBytes = maxTokenCountSingleItemPayloadSize(session, request)

	response, err := getTextEmbeddingNumTokensServerlessAware(session, request)
	require.NoError(t, err)
	readAllStreamValues(t, response)

	require.NotContains(t, logOutput.String(), sensitiveText)
	require.NotContains(t, logOutput.String(), sensitiveCredential)
}

func TestInvokeTextEmbeddingServerlessAwareMergesResultsAndUsage(t *testing.T) {
	originalTexts := []string{"a", "bb", "ccc"}
	request := newTextEmbeddingRequest(originalTexts)
	runtime := &embeddingBatchRuntime{}
	session := newTestSession(
		t,
		runtime,
		access_types.PLUGIN_ACCESS_TYPE_MODEL,
		access_types.PLUGIN_ACCESS_ACTION_INVOKE_TEXT_EMBEDDING,
	)
	runtime.maxRequestBytes = maxTextEmbeddingSingleItemPayloadSize(session, request)

	response, err := invokeTextEmbeddingServerlessAware(session, request)
	require.NoError(t, err)
	results := readAllStreamValues(t, response)

	require.Len(t, results, 1)
	result := results[0]
	require.Equal(t, "embedding-model", result.Model)
	require.Equal(t, [][]float64{{1}, {2}, {3}}, result.Embeddings)
	require.Equal(t, 3, *result.Usage.Tokens)
	require.Equal(t, 30, *result.Usage.TotalTokens)
	require.True(t, decimal.RequireFromString("0.01").Equal(result.Usage.UnitPrice))
	require.True(t, decimal.NewFromInt(1000).Equal(result.Usage.PriceUnit))
	require.True(t, decimal.RequireFromString("0.06").Equal(result.Usage.TotalPrice))
	require.Equal(t, "USD", *result.Usage.Currency)
	require.Equal(t, 1.5, *result.Usage.Latency)
	require.Equal(t, originalTexts, request.Texts)

	runtime.mu.Lock()
	defer runtime.mu.Unlock()
	require.Len(t, runtime.payloads, 3)
	require.Equal(t, 1, runtime.maxInFlight)
	for _, payload := range runtime.payloads {
		require.LessOrEqual(t, len(payload), runtime.maxRequestBytes)
	}
}

func TestInvokeTextEmbeddingServerlessAwareRejectsInconsistentBatchUsage(t *testing.T) {
	tests := []struct {
		name   string
		reason string
		mutate func(*model_entities.TextEmbeddingResult)
	}{
		{
			name:   "model",
			reason: "model",
			mutate: func(result *model_entities.TextEmbeddingResult) {
				result.Model = "different-model"
			},
		},
		{
			name:   "unit price",
			reason: "unit_price",
			mutate: func(result *model_entities.TextEmbeddingResult) {
				result.Usage.UnitPrice = decimal.RequireFromString("0.02")
			},
		},
		{
			name:   "price unit",
			reason: "price_unit",
			mutate: func(result *model_entities.TextEmbeddingResult) {
				result.Usage.PriceUnit = decimal.NewFromInt(1)
			},
		},
		{
			name:   "currency",
			reason: "currency",
			mutate: func(result *model_entities.TextEmbeddingResult) {
				currency := "EUR"
				result.Usage.Currency = &currency
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			request := newTextEmbeddingRequest([]string{"a", "bb", "ccc"})
			runtime := &embeddingBatchRuntime{
				embeddingResponseFn: func(callIndex int, texts []string) []plugin_entities.SessionMessage {
					result := embeddingResult(texts)
					if callIndex == 1 {
						tt.mutate(&result)
					}
					return []plugin_entities.SessionMessage{
						{
							Type: plugin_entities.SESSION_MESSAGE_TYPE_STREAM,
							Data: parser.MarshalJsonBytes(result),
						},
						{Type: plugin_entities.SESSION_MESSAGE_TYPE_END},
					}
				},
			}
			session := newTestSession(
				t,
				runtime,
				access_types.PLUGIN_ACCESS_TYPE_MODEL,
				access_types.PLUGIN_ACCESS_ACTION_INVOKE_TEXT_EMBEDDING,
			)
			runtime.maxRequestBytes = maxTextEmbeddingSingleItemPayloadSize(session, request)

			response, err := invokeTextEmbeddingServerlessAware(session, request)
			require.NoError(t, err)
			values, streamErr := collectTextEmbeddingStream(response)

			require.ErrorContains(t, streamErr, "ServerlessBatchResultMismatch")
			require.ErrorContains(t, streamErr, "reason="+tt.reason)
			require.Empty(t, values)
			runtime.mu.Lock()
			defer runtime.mu.Unlock()
			require.Len(t, runtime.payloads, 2)
		})
	}
}

func TestInvokeTextEmbeddingServerlessAwareRejectsEmbeddingCountMismatch(t *testing.T) {
	request := newTextEmbeddingRequest([]string{"a", "bb"})
	runtime := &embeddingBatchRuntime{
		embeddingResponseFn: func(_ int, texts []string) []plugin_entities.SessionMessage {
			result := embeddingResult(texts)
			result.Embeddings = append(result.Embeddings, []float64{999})
			return []plugin_entities.SessionMessage{
				{
					Type: plugin_entities.SESSION_MESSAGE_TYPE_STREAM,
					Data: parser.MarshalJsonBytes(result),
				},
				{Type: plugin_entities.SESSION_MESSAGE_TYPE_END},
			}
		},
	}
	session := newTestSession(
		t,
		runtime,
		access_types.PLUGIN_ACCESS_TYPE_MODEL,
		access_types.PLUGIN_ACCESS_ACTION_INVOKE_TEXT_EMBEDDING,
	)
	runtime.maxRequestBytes = maxTextEmbeddingSingleItemPayloadSize(session, request)

	response, err := invokeTextEmbeddingServerlessAware(session, request)
	require.NoError(t, err)
	values, streamErr := collectTextEmbeddingStream(response)

	require.ErrorContains(t, streamErr, "reason=embedding_count")
	require.Empty(t, values)
}

func BenchmarkSplitTokenCountRequest100MiB(b *testing.B) {
	const (
		requestBytes = 100 * 1024 * 1024
		textBytes    = 1024 * 1024
	)

	text := strings.Repeat("a", textBytes)
	texts := make([]string, requestBytes/textBytes)
	for i := range texts {
		texts[i] = text
	}
	request := newTokenCountRequest(texts)
	runtime := &tokenBatchRuntime{
		maxRequestBytes: 5 * 1024 * 1024,
	}
	session := session_manager.NewSession(session_manager.NewSessionPayload{
		TenantID:               "tenant-1",
		UserID:                 "user-1",
		PluginUniqueIdentifier: testPluginUniqueIdentifier,
		InvokeFrom:             access_types.PLUGIN_ACCESS_TYPE_MODEL,
		Action:                 access_types.PLUGIN_ACCESS_ACTION_GET_TEXT_EMBEDDING_NUM_TOKENS,
		RequestContext:         context.Background(),
		IgnoreCache:            true,
	})
	session.BindRuntime(runtime)
	b.Cleanup(func() {
		session.Close(session_manager.CloseSessionPayload{IgnoreCache: true})
	})

	b.ReportAllocs()
	b.SetBytes(requestBytes)
	for b.Loop() {
		batches, err := splitTokenCountRequest(session, request, runtime.maxRequestBytes)
		if err != nil {
			b.Fatal(err)
		}
		if len(batches) == 0 {
			b.Fatal("expected split batches")
		}
	}
}

func newTokenCountRequest(texts []string) *requests.RequestGetTextEmbeddingNumTokens {
	return &requests.RequestGetTextEmbeddingNumTokens{
		BaseRequestInvokeModel: requests.BaseRequestInvokeModel{
			Provider: "provider",
			Model:    "model",
		},
		Credentials: requests.Credentials{
			Credentials: map[string]any{"api_key": "secret"},
		},
		ModelType: model_entities.MODEL_TYPE_TEXT_EMBEDDING,
		Texts:     texts,
	}
}

func newTextEmbeddingRequest(texts []string) *requests.RequestInvokeTextEmbedding {
	return &requests.RequestInvokeTextEmbedding{
		BaseRequestInvokeModel: requests.BaseRequestInvokeModel{
			Provider: "provider",
			Model:    "model",
		},
		Credentials: requests.Credentials{
			Credentials: map[string]any{"api_key": "secret"},
		},
		InvokeTextEmbeddingSchema: requests.InvokeTextEmbeddingSchema{
			Texts:     texts,
			InputType: "document",
		},
		ModelType: model_entities.MODEL_TYPE_TEXT_EMBEDDING,
	}
}

func tokenCountPayloadSize(
	session *session_manager.Session,
	request *requests.RequestGetTextEmbeddingNumTokens,
) int {
	return len(session.Message(
		session_manager.PLUGIN_IN_STREAM_EVENT_REQUEST,
		GetInvokePluginMap(session, request),
	))
}

func maxTokenCountSingleItemPayloadSize(
	session *session_manager.Session,
	request *requests.RequestGetTextEmbeddingNumTokens,
) int {
	maxPayloadBytes := 0
	for i := range request.Texts {
		singleItemRequest := *request
		singleItemRequest.Texts = request.Texts[i : i+1]
		maxPayloadBytes = max(maxPayloadBytes, tokenCountPayloadSize(session, &singleItemRequest))
	}
	return maxPayloadBytes
}

func maxTextEmbeddingSingleItemPayloadSize(
	session *session_manager.Session,
	request *requests.RequestInvokeTextEmbedding,
) int {
	maxPayloadBytes := 0
	for i := range request.Texts {
		singleItemRequest := *request
		singleItemRequest.Texts = request.Texts[i : i+1]
		maxPayloadBytes = max(
			maxPayloadBytes,
			len(session.Message(
				session_manager.PLUGIN_IN_STREAM_EVENT_REQUEST,
				GetInvokePluginMap(session, &singleItemRequest),
			)),
		)
	}
	return maxPayloadBytes
}

func readAllStreamValues[T any](t *testing.T, response interface {
	Next() bool
	Read() (T, error)
}) []T {
	t.Helper()

	var values []T
	for response.Next() {
		value, err := response.Read()
		require.NoError(t, err)
		values = append(values, value)
	}
	return values
}

func collectTokenCountStream(
	response interface {
		Next() bool
		Read() (model_entities.GetTextEmbeddingNumTokensResponse, error)
	},
) ([]model_entities.GetTextEmbeddingNumTokensResponse, error) {
	var values []model_entities.GetTextEmbeddingNumTokensResponse
	for response.Next() {
		value, err := response.Read()
		if err != nil {
			return values, err
		}
		values = append(values, value)
	}
	return values, nil
}

func collectTextEmbeddingStream(
	response interface {
		Next() bool
		Read() (model_entities.TextEmbeddingResult, error)
	},
) ([]model_entities.TextEmbeddingResult, error) {
	var values []model_entities.TextEmbeddingResult
	for response.Next() {
		value, err := response.Read()
		if err != nil {
			return values, err
		}
		values = append(values, value)
	}
	return values, nil
}

func int64HistogramSummary(
	metrics metricdata.ResourceMetrics,
	metricName string,
) (uint64, int64, bool) {
	var count uint64
	var sum int64
	found := false
	for _, scopeMetrics := range metrics.ScopeMetrics {
		for _, currentMetric := range scopeMetrics.Metrics {
			if currentMetric.Name != metricName {
				continue
			}
			histogram, ok := currentMetric.Data.(metricdata.Histogram[int64])
			if !ok {
				return 0, 0, false
			}
			found = true
			for _, point := range histogram.DataPoints {
				count += point.Count
				sum += point.Sum
			}
		}
	}
	return count, sum, found
}

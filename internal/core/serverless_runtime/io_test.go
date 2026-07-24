package serverless_runtime

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	"github.com/langgenius/dify-plugin-daemon/internal/core/io_tunnel/access_types"
	"github.com/langgenius/dify-plugin-daemon/pkg/entities"
	"github.com/langgenius/dify-plugin-daemon/pkg/entities/plugin_entities"
	"github.com/langgenius/dify-plugin-daemon/pkg/utils/mapping"
	"github.com/langgenius/dify-plugin-daemon/pkg/utils/routine"
)

func TestShouldRetryStatusCode(t *testing.T) {
	tests := []struct {
		name       string
		statusCode int
		expected   bool
	}{
		{"502 should retry", 502, true},
		{"200 should not retry", 200, false},
		{"404 should not retry", 404, false},
		{"429 should not retry", 429, false},
		{"500 should not retry", 500, false},
		{"503 should not retry", 503, false},
		{"504 should not retry", 504, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := shouldRetryStatusCode(tt.statusCode)
			if result != tt.expected {
				t.Errorf("shouldRetryStatusCode(%d) = %v, expected %v", tt.statusCode, result, tt.expected)
			}
		})
	}
}

func TestServerlessRuntimeFailureType(t *testing.T) {
	tests := []struct {
		statusCode int
		expected   string
	}{
		{statusCode: 0, expected: "transport"},
		{statusCode: http.StatusTooManyRequests, expected: "rate_limit"},
		{statusCode: http.StatusBadGateway, expected: "gateway"},
		{statusCode: http.StatusInternalServerError, expected: "http"},
	}

	for _, tt := range tests {
		if actual := serverlessRuntimeFailureType(tt.statusCode); actual != tt.expected {
			t.Errorf(
				"serverlessRuntimeFailureType(%d) = %q, expected %q",
				tt.statusCode,
				actual,
				tt.expected,
			)
		}
	}
}

func TestListenRemovesClosedListener(t *testing.T) {
	runtime := &ServerlessPluginRuntime{
		listeners: mapping.Map[string, *entities.Broadcast[plugin_entities.SessionMessage]]{},
	}
	listener, err := runtime.Listen("test-session")
	if err != nil {
		t.Fatalf("listen failed: %v", err)
	}

	var externalCloseCalls atomic.Int32
	listener.OnClose(func() {
		externalCloseCalls.Add(1)
	})

	listener.Close()
	listener.Close()

	if runtime.listeners.Len() != 0 {
		t.Fatalf("expected listener cleanup, got %d listeners", runtime.listeners.Len())
	}
	if got := externalCloseCalls.Load(); got != 1 {
		t.Fatalf("expected external close callback once, got %d calls", got)
	}
}

func TestInvokeServerlessWithRetry_Success(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("success"))
	}))
	defer server.Close()

	runtime := &ServerlessPluginRuntime{
		Client:                    server.Client(),
		MaxRetryTimes:             3,
		PluginMaxExecutionTimeout: 10,
	}

	response, err := runtime.invokeServerlessWithRetry(
		context.Background(),
		server.URL,
		"test-session",
		[]byte("test-data"),
		"",
	)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if response.StatusCode != http.StatusOK {
		t.Errorf("Expected status 200, got %d", response.StatusCode)
	}

	body, _ := io.ReadAll(response.Body)
	response.Body.Close()
	if string(body) != "success" {
		t.Errorf("Expected body 'success', got '%s'", string(body))
	}
}

func TestInvokeServerlessWithRetry_RetryOn502(t *testing.T) {
	attemptCount := atomic.Int32{}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		attempt := attemptCount.Add(1)
		if attempt < 3 {
			w.WriteHeader(http.StatusBadGateway)
			w.Write([]byte("bad gateway"))
		} else {
			w.WriteHeader(http.StatusOK)
			w.Write([]byte("success"))
		}
	}))
	defer server.Close()

	runtime := &ServerlessPluginRuntime{
		Client:                    server.Client(),
		MaxRetryTimes:             3,
		PluginMaxExecutionTimeout: 10,
	}

	startTime := time.Now()
	response, err := runtime.invokeServerlessWithRetry(
		context.Background(),
		server.URL,
		"test-session",
		[]byte("test-data"),
		"",
	)
	duration := time.Since(startTime)

	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if response.StatusCode != http.StatusOK {
		t.Errorf("Expected status 200, got %d", response.StatusCode)
	}

	if attemptCount.Load() != 3 {
		t.Errorf("Expected 3 attempts, got %d", attemptCount.Load())
	}

	expectedMinDuration := 500*time.Millisecond + 1000*time.Millisecond
	if duration < expectedMinDuration {
		t.Errorf("Expected at least %v duration for backoff, got %v", expectedMinDuration, duration)
	}

	body, _ := io.ReadAll(response.Body)
	response.Body.Close()
	if string(body) != "success" {
		t.Errorf("Expected body 'success', got '%s'", string(body))
	}
}

func TestInvokeServerlessWithRetry_NoRetryOn404(t *testing.T) {
	attemptCount := atomic.Int32{}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		attemptCount.Add(1)
		w.WriteHeader(http.StatusNotFound)
		w.Write([]byte("not found"))
	}))
	defer server.Close()

	runtime := &ServerlessPluginRuntime{
		Client:                    server.Client(),
		MaxRetryTimes:             3,
		PluginMaxExecutionTimeout: 10,
	}

	response, err := runtime.invokeServerlessWithRetry(
		context.Background(),
		server.URL,
		"test-session",
		[]byte("test-data"),
		"",
	)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if response.StatusCode != http.StatusNotFound {
		t.Errorf("Expected status 404, got %d", response.StatusCode)
	}

	if attemptCount.Load() != 1 {
		t.Errorf("Expected 1 attempt (no retry), got %d", attemptCount.Load())
	}
}

func TestInvokeServerlessWithRetry_NoRetryOn500(t *testing.T) {
	attemptCount := atomic.Int32{}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		attemptCount.Add(1)
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("internal server error"))
	}))
	defer server.Close()

	runtime := &ServerlessPluginRuntime{
		Client:                    server.Client(),
		MaxRetryTimes:             3,
		PluginMaxExecutionTimeout: 10,
	}

	response, err := runtime.invokeServerlessWithRetry(
		context.Background(),
		server.URL,
		"test-session",
		[]byte("test-data"),
		"",
	)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if response.StatusCode != http.StatusInternalServerError {
		t.Errorf("Expected status 500, got %d", response.StatusCode)
	}

	if attemptCount.Load() != 1 {
		t.Errorf("Expected 1 attempt (no retry), got %d", attemptCount.Load())
	}
}

func TestInvokeServerlessWithRetry_MaxRetriesExceeded(t *testing.T) {
	attemptCount := atomic.Int32{}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		attemptCount.Add(1)
		w.WriteHeader(http.StatusBadGateway)
		w.Write([]byte("bad gateway"))
	}))
	defer server.Close()

	runtime := &ServerlessPluginRuntime{
		Client:                    server.Client(),
		MaxRetryTimes:             3,
		PluginMaxExecutionTimeout: 10,
	}

	response, err := runtime.invokeServerlessWithRetry(
		context.Background(),
		server.URL,
		"test-session",
		[]byte("test-data"),
		"",
	)

	if err == nil {
		t.Fatal("Expected error after max retries, got nil")
	}

	if response != nil {
		t.Errorf("Expected nil response after max retries, got %v", response)
	}

	if attemptCount.Load() != 3 {
		t.Errorf("Expected 3 attempts, got %d", attemptCount.Load())
	}

	expectedError := "all 3 attempts failed, last error: attempt 3/3 failed with status code: 502"
	if err.Error()[:len(expectedError)] != expectedError {
		t.Errorf("Expected error message to start with '%s', got '%s'", expectedError, err.Error())
	}
}

func TestInvokeServerlessWithRetry_ExponentialBackoff(t *testing.T) {
	attemptTimes := make([]time.Time, 0)

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		attemptTimes = append(attemptTimes, time.Now())
		if len(attemptTimes) < 3 {
			w.WriteHeader(http.StatusBadGateway)
		} else {
			w.WriteHeader(http.StatusOK)
		}
	}))
	defer server.Close()

	runtime := &ServerlessPluginRuntime{
		Client:                    server.Client(),
		MaxRetryTimes:             3,
		PluginMaxExecutionTimeout: 10,
	}

	_, err := runtime.invokeServerlessWithRetry(
		context.Background(),
		server.URL,
		"test-session",
		[]byte("test-data"),
		"",
	)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if len(attemptTimes) != 3 {
		t.Fatalf("Expected 3 attempts, got %d", len(attemptTimes))
	}

	backoff1 := attemptTimes[1].Sub(attemptTimes[0])
	backoff2 := attemptTimes[2].Sub(attemptTimes[1])

	minBackoff1 := 500 * time.Millisecond
	minBackoff2 := 1000 * time.Millisecond

	if backoff1 < minBackoff1 {
		t.Errorf("First backoff should be at least %v, got %v", minBackoff1, backoff1)
	}

	if backoff2 < minBackoff2 {
		t.Errorf("Second backoff should be at least %v, got %v", minBackoff2, backoff2)
	}

	if backoff2 <= backoff1 {
		t.Errorf("Backoff should be exponential: second (%v) should be greater than first (%v)", backoff2, backoff1)
	}
}

func TestInvokeServerlessWithRetry_CancellationStopsBackoff(t *testing.T) {
	requestCount := atomic.Int32{}
	requestReceived := make(chan struct{}, 1)
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requestCount.Add(1)
		requestReceived <- struct{}{}
		w.WriteHeader(http.StatusBadGateway)
	}))
	defer server.Close()

	runtime := &ServerlessPluginRuntime{
		Client:                    server.Client(),
		MaxRetryTimes:             3,
		PluginMaxExecutionTimeout: 10,
	}
	ctx, cancel := context.WithCancel(context.Background())
	result := make(chan error, 1)
	go func() {
		response, err := runtime.invokeServerlessWithRetry(
			ctx,
			server.URL,
			"test-session",
			[]byte("test-data"),
			"",
		)
		if response != nil {
			response.Body.Close()
		}
		result <- err
	}()

	<-requestReceived
	cancel()
	select {
	case err := <-result:
		if !errors.Is(err, context.Canceled) {
			t.Fatalf("expected context cancellation, got %v", err)
		}
	case <-time.After(time.Second):
		t.Fatal("retry backoff did not stop after cancellation")
	}
	if requestCount.Load() != 1 {
		t.Fatalf("expected one request before cancellation, got %d", requestCount.Load())
	}
}

func TestInvokeServerlessWithRetry_MaxRetriesZero(t *testing.T) {
	attemptCount := atomic.Int32{}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		attemptCount.Add(1)
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	runtime := &ServerlessPluginRuntime{
		Client:                    server.Client(),
		MaxRetryTimes:             0,
		PluginMaxExecutionTimeout: 10,
	}

	response, err := runtime.invokeServerlessWithRetry(
		context.Background(),
		server.URL,
		"test-session",
		[]byte("test-data"),
		"",
	)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if response.StatusCode != http.StatusOK {
		t.Errorf("Expected status 200, got %d", response.StatusCode)
	}

	if attemptCount.Load() != 1 {
		t.Errorf("Expected 1 attempt even with MaxRetryTimes=0, got %d", attemptCount.Load())
	}
}

func TestInvokeServerlessWithRetry_RequestData(t *testing.T) {
	receivedData := ""

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, _ := io.ReadAll(r.Body)
		receivedData = string(body)

		if r.Header.Get("Dify-Plugin-Session-ID") != "test-session-123" {
			t.Errorf("Expected session ID 'test-session-123', got '%s'", r.Header.Get("Dify-Plugin-Session-ID"))
		}

		if r.Header.Get("Content-Type") != "application/json" {
			t.Errorf("Expected Content-Type 'application/json', got '%s'", r.Header.Get("Content-Type"))
		}

		if r.Header.Get("Accept") != "text/event-stream" {
			t.Errorf("Expected Accept 'text/event-stream', got '%s'", r.Header.Get("Accept"))
		}

		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	runtime := &ServerlessPluginRuntime{
		Client:                    server.Client(),
		MaxRetryTimes:             3,
		PluginMaxExecutionTimeout: 10,
	}

	testData := []byte(`{"test": "data"}`)
	_, err := runtime.invokeServerlessWithRetry(
		context.Background(),
		server.URL,
		"test-session-123",
		testData,
		"",
	)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if receivedData != string(testData) {
		t.Errorf("Expected received data '%s', got '%s'", string(testData), receivedData)
	}
}

func TestListen(t *testing.T) {
	runtime := &ServerlessPluginRuntime{
		listeners: mapping.Map[string, *entities.Broadcast[plugin_entities.SessionMessage]]{},
	}

	sessionID := "test-session"
	broadcast, err := runtime.Listen(sessionID)

	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if broadcast == nil {
		t.Fatal("Expected broadcast to be non-nil")
	}

	stored, ok := runtime.listeners.Load(sessionID)
	if !ok {
		t.Fatal("Expected listener to be stored")
	}

	if stored != broadcast {
		t.Error("Stored listener should match returned broadcast")
	}
}

func TestWrite_PayloadAtLimitIsSent(t *testing.T) {
	requestCount := atomic.Int32{}
	payload := []byte("1234")
	received := collectServerlessWriteMessagesWithPayload(
		t,
		func(w http.ResponseWriter, r *http.Request) {
			requestCount.Add(1)
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte(
				`{"session_id":"test-session","event":"session","data":{"type":"stream","data":{"result":true}}}`,
			))
		},
		len(payload),
		payload,
	)

	if requestCount.Load() != 1 {
		t.Fatalf("expected one HTTP request, got %d", requestCount.Load())
	}
	if len(received) != 2 {
		t.Fatalf("expected stream and end messages, got %d: %#v", len(received), received)
	}
}

func TestWriteTreatsPluginEndAsTerminal(t *testing.T) {
	routine.InitPool(1)
	requestCanceled := make(chan struct{})
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(
			"{\"session_id\":\"test-session\",\"event\":\"session\",\"data\":{\"type\":\"end\",\"data\":{\"reason\":\"done\"}}}\n" +
				"{\"session_id\":\"test-session\",\"event\":\"session\",\"data\":{\"type\":\"stream\",\"data\":{\"result\":true}}}\n",
		))
		w.(http.Flusher).Flush()
		<-r.Context().Done()
		close(requestCanceled)
	}))
	t.Cleanup(server.Close)

	runtime := &ServerlessPluginRuntime{
		Client:                    server.Client(),
		LambdaURL:                 server.URL,
		MaxRequestBytes:           1024,
		MaxRetryTimes:             1,
		PluginMaxExecutionTimeout: 10,
		RuntimeBufferSize:         1024,
		RuntimeMaxBufferSize:      1024,
		listeners:                 mapping.Map[string, *entities.Broadcast[plugin_entities.SessionMessage]]{},
	}
	listener, err := runtime.Listen("test-session")
	if err != nil {
		t.Fatalf("listen failed: %v", err)
	}

	streamSeen := make(chan struct{}, 1)
	endSeen := make(chan json.RawMessage, 1)
	listener.Listen(func(message plugin_entities.SessionMessage) {
		switch message.Type {
		case plugin_entities.SESSION_MESSAGE_TYPE_STREAM:
			streamSeen <- struct{}{}
		case plugin_entities.SESSION_MESSAGE_TYPE_END:
			endSeen <- message.Data
		}
	})

	if err := runtime.Write(
		"test-session",
		access_types.PLUGIN_ACCESS_ACTION_INVOKE_TEXT_EMBEDDING,
		[]byte("{}"),
	); err != nil {
		t.Fatalf("write failed: %v", err)
	}

	select {
	case endData := <-endSeen:
		if string(endData) != `{"reason":"done"}` {
			t.Fatalf("expected plugin END data to be preserved, got %s", endData)
		}
	case <-time.After(time.Second):
		t.Fatal("timed out waiting for terminal end")
	}
	select {
	case <-streamSeen:
		t.Fatal("received stream event after terminal end")
	default:
	}
	select {
	case <-requestCanceled:
	case <-time.After(time.Second):
		t.Fatal("server request was not canceled after terminal end")
	}
	if runtime.listeners.Len() != 0 {
		t.Fatalf("expected listener cleanup, got %d listeners", runtime.listeners.Len())
	}
}

func TestWriteContextCancelsHTTPRequest(t *testing.T) {
	routine.InitPool(1)
	requestStarted := make(chan struct{})
	requestCanceled := make(chan struct{})
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		close(requestStarted)
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(
			"{\"session_id\":\"test-session\",\"event\":\"session\",\"data\":{\"type\":\"stream\",\"data\":{\"result\":true}}}\n",
		))
		w.(http.Flusher).Flush()
		<-r.Context().Done()
		close(requestCanceled)
	}))
	t.Cleanup(server.Close)

	runtime := &ServerlessPluginRuntime{
		Client:                    server.Client(),
		LambdaURL:                 server.URL,
		MaxRequestBytes:           1024,
		MaxRetryTimes:             1,
		PluginMaxExecutionTimeout: 10,
		RuntimeBufferSize:         1024,
		RuntimeMaxBufferSize:      1024,
		listeners:                 mapping.Map[string, *entities.Broadcast[plugin_entities.SessionMessage]]{},
	}
	listener, err := runtime.Listen("test-session")
	if err != nil {
		t.Fatalf("listen failed: %v", err)
	}
	streamSeen := make(chan struct{}, 1)
	listenerClosed := make(chan struct{})
	listener.Listen(func(message plugin_entities.SessionMessage) {
		if message.Type == plugin_entities.SESSION_MESSAGE_TYPE_STREAM {
			streamSeen <- struct{}{}
		}
	})
	listener.OnClose(func() {
		close(listenerClosed)
	})

	ctx, cancel := context.WithCancel(context.Background())
	if err := runtime.WriteContext(
		ctx,
		"test-session",
		access_types.PLUGIN_ACCESS_ACTION_INVOKE_TEXT_EMBEDDING,
		[]byte("{}"),
	); err != nil {
		t.Fatalf("write failed: %v", err)
	}
	<-requestStarted
	select {
	case <-streamSeen:
	case <-time.After(time.Second):
		t.Fatal("timed out waiting for stream event")
	}

	cancel()
	select {
	case <-requestCanceled:
	case <-time.After(time.Second):
		t.Fatal("server request did not observe cancellation")
	}
	select {
	case <-listenerClosed:
	case <-time.After(time.Second):
		t.Fatal("listener was not closed after cancellation")
	}
	if runtime.listeners.Len() != 0 {
		t.Fatalf("expected listener cleanup, got %d listeners", runtime.listeners.Len())
	}
}

func TestWrite_ExecutionTimeoutSendsRuntimeError(t *testing.T) {
	routine.InitPool(1)
	requestCanceled := make(chan struct{})
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.(http.Flusher).Flush()
		<-r.Context().Done()
		close(requestCanceled)
	}))
	t.Cleanup(server.Close)

	runtime := &ServerlessPluginRuntime{
		Client:                    server.Client(),
		LambdaURL:                 server.URL,
		MaxRequestBytes:           1024,
		MaxRetryTimes:             1,
		PluginMaxExecutionTimeout: 1,
		RuntimeBufferSize:         1024,
		RuntimeMaxBufferSize:      1024,
		listeners:                 mapping.Map[string, *entities.Broadcast[plugin_entities.SessionMessage]]{},
	}
	listener, err := runtime.Listen("test-session")
	if err != nil {
		t.Fatalf("listen failed: %v", err)
	}
	messages := make(chan plugin_entities.SessionMessage, 2)
	listenerClosed := make(chan struct{})
	listener.Listen(func(message plugin_entities.SessionMessage) {
		messages <- message
	})
	listener.OnClose(func() {
		close(listenerClosed)
	})

	if err := runtime.Write(
		"test-session",
		access_types.PLUGIN_ACCESS_ACTION_INVOKE_TEXT_EMBEDDING,
		[]byte("{}"),
	); err != nil {
		t.Fatalf("write failed: %v", err)
	}
	select {
	case <-listenerClosed:
	case <-time.After(2 * time.Second):
		t.Fatal("listener was not closed after execution timeout")
	}
	select {
	case <-requestCanceled:
	case <-time.After(time.Second):
		t.Fatal("server request did not observe execution timeout")
	}

	close(messages)
	received := make([]plugin_entities.SessionMessage, 0, len(messages))
	for message := range messages {
		received = append(received, message)
	}
	if len(received) != 1 || received[0].Type != plugin_entities.SESSION_MESSAGE_TYPE_ERROR {
		t.Fatalf("expected one runtime error, got %#v", received)
	}
	var errorResponse plugin_entities.ErrorResponse
	if err := json.Unmarshal(received[0].Data, &errorResponse); err != nil {
		t.Fatalf("unmarshal error response: %v", err)
	}
	if !strings.Contains(errorResponse.Message, context.DeadlineExceeded.Error()) {
		t.Fatalf("expected deadline error, got %q", errorResponse.Message)
	}
}

func TestWrite_PayloadOverLimitFailsBeforeHTTPRequest(t *testing.T) {
	requestCount := atomic.Int32{}
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requestCount.Add(1)
		w.WriteHeader(http.StatusOK)
	}))
	t.Cleanup(server.Close)

	runtime := &ServerlessPluginRuntime{
		Client:                    server.Client(),
		LambdaURL:                 server.URL,
		MaxRequestBytes:           3,
		MaxRetryTimes:             3,
		PluginMaxExecutionTimeout: 10,
		RuntimeBufferSize:         1024,
		RuntimeMaxBufferSize:      1024,
		listeners:                 mapping.Map[string, *entities.Broadcast[plugin_entities.SessionMessage]]{},
	}
	listener, err := runtime.Listen("test-session")
	if err != nil {
		t.Fatalf("listen failed: %v", err)
	}
	t.Cleanup(listener.Close)

	action := access_types.PLUGIN_ACCESS_ACTION_VALIDATE_TOOL_CREDENTIALS
	err = runtime.Write("test-session", action, []byte("1234"))
	if err == nil {
		t.Fatal("expected oversized payload to fail")
	}

	var payloadTooLargeError *ServerlessPayloadTooLargeError
	if !errors.As(err, &payloadTooLargeError) {
		t.Fatalf("expected ServerlessPayloadTooLargeError, got %T: %v", err, err)
	}
	if payloadTooLargeError.Action != action {
		t.Errorf("expected action %q, got %q", action, payloadTooLargeError.Action)
	}
	if payloadTooLargeError.PayloadBytes != 4 {
		t.Errorf("expected payload size 4, got %d", payloadTooLargeError.PayloadBytes)
	}
	if payloadTooLargeError.MaxRequestBytes != 3 {
		t.Errorf("expected max request size 3, got %d", payloadTooLargeError.MaxRequestBytes)
	}
	if requestCount.Load() != 0 {
		t.Fatalf("expected no HTTP requests, got %d", requestCount.Load())
	}
	if runtime.listeners.Len() != 0 {
		t.Fatalf("expected listener cleanup, got %d listeners", runtime.listeners.Len())
	}
}

func TestWrite_NonSuccessResponseSendsRuntimeError(t *testing.T) {
	received := collectServerlessWriteMessages(t, func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("x-amzn-ErrorType", "Runtime.ExitError")
		w.Header().Set("x-amzn-RequestId", "lambda-request-id")
		w.WriteHeader(http.StatusInternalServerError)
		_, _ = w.Write([]byte(`{"errorType":"Runtime.ExitError"}`))
	})

	requireRuntimeError(
		t,
		received,
		"Plugin runtime request failed: Runtime.ExitError",
		http.StatusInternalServerError,
	)
}

func TestWrite_RateLimitDoesNotRetry(t *testing.T) {
	requestCount := atomic.Int32{}
	received := collectServerlessWriteMessages(t, func(w http.ResponseWriter, _ *http.Request) {
		requestCount.Add(1)
		w.Header().Set("x-amzn-ErrorType", "TooManyRequestsException")
		w.Header().Set("x-amzn-RequestId", "lambda-request-id")
		w.WriteHeader(http.StatusTooManyRequests)
		_, _ = w.Write([]byte(`{"errorType":"TooManyRequestsException","errorMessage":"rate exceeded"}`))
	})

	if requestCount.Load() != 1 {
		t.Fatalf("expected 429 response not to be retried, got %d requests", requestCount.Load())
	}
	requireRuntimeError(
		t,
		received,
		"Plugin runtime request failed: TooManyRequestsException: rate exceeded",
		http.StatusTooManyRequests,
	)
}

func TestWrite_SuccessResponseWithLambdaErrorBodySendsRuntimeError(t *testing.T) {
	received := collectServerlessWriteMessages(t, func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("x-amzn-RequestId", "lambda-request-id")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"errorType":"Runtime.ExitError","errorMessage":"Runtime exited"}`))
	})

	requireRuntimeError(
		t,
		received,
		"Plugin runtime request failed: Runtime.ExitError: Runtime exited",
		http.StatusOK,
	)
}

func TestWrite_EmptySuccessResponseSendsRuntimeError(t *testing.T) {
	received := collectServerlessWriteMessages(t, func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("x-amzn-RequestId", "lambda-request-id")
		w.WriteHeader(http.StatusOK)
	})

	requireRuntimeError(
		t,
		received,
		"Plugin runtime request failed: no valid session response",
		http.StatusOK,
	)
}

func TestWrite_SuccessResponseWithSessionEventStillEndsNormally(t *testing.T) {
	received := collectServerlessWriteMessages(t, func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(
			`{"session_id":"test-session","event":"session","data":{"type":"stream","data":{"result":true}}}`,
		))
	})

	if len(received) != 2 {
		t.Fatalf("expected stream and end messages, got %d: %#v", len(received), received)
	}
	if received[0].Type != plugin_entities.SESSION_MESSAGE_TYPE_STREAM {
		t.Fatalf("expected stream message, got %s", received[0].Type)
	}
	if received[1].Type != plugin_entities.SESSION_MESSAGE_TYPE_END {
		t.Fatalf("expected end message, got %s", received[1].Type)
	}
}

func collectServerlessWriteMessages(t *testing.T, handler http.HandlerFunc) []plugin_entities.SessionMessage {
	return collectServerlessWriteMessagesWithPayload(
		t,
		handler,
		1024,
		[]byte(`{"credentials":{"api_key":"secret"}}`),
	)
}

func collectServerlessWriteMessagesWithPayload(
	t *testing.T,
	handler http.HandlerFunc,
	maxRequestBytes int,
	payload []byte,
) []plugin_entities.SessionMessage {
	t.Helper()
	routine.InitPool(1)

	server := httptest.NewServer(handler)
	t.Cleanup(server.Close)

	runtime := &ServerlessPluginRuntime{
		Client:                    server.Client(),
		LambdaURL:                 server.URL,
		MaxRequestBytes:           maxRequestBytes,
		MaxRetryTimes:             1,
		PluginMaxExecutionTimeout: 10,
		RuntimeBufferSize:         1024,
		RuntimeMaxBufferSize:      1024,
		listeners:                 mapping.Map[string, *entities.Broadcast[plugin_entities.SessionMessage]]{},
	}

	listener, err := runtime.Listen("test-session")
	if err != nil {
		t.Fatalf("listen failed: %v", err)
	}

	messages := make(chan plugin_entities.SessionMessage, 4)
	done := make(chan struct{})
	listener.Listen(func(message plugin_entities.SessionMessage) {
		messages <- message
	})
	listener.OnClose(func() {
		close(done)
	})

	if err := runtime.Write(
		"test-session",
		access_types.PLUGIN_ACCESS_ACTION_VALIDATE_TOOL_CREDENTIALS,
		payload,
	); err != nil {
		t.Fatalf("write failed: %v", err)
	}

	select {
	case <-done:
	case <-time.After(time.Second):
		t.Fatal("timed out waiting for serverless response")
	}
	close(messages)

	received := make([]plugin_entities.SessionMessage, 0, len(messages))
	for message := range messages {
		received = append(received, message)
	}
	return received
}

func requireRuntimeError(
	t *testing.T,
	received []plugin_entities.SessionMessage,
	wantMessage string,
	wantStatusCode int,
) {
	t.Helper()
	if len(received) != 1 {
		t.Fatalf("expected one terminal message, got %d: %#v", len(received), received)
	}
	if received[0].Type != plugin_entities.SESSION_MESSAGE_TYPE_ERROR {
		t.Fatalf("expected runtime error, got %s", received[0].Type)
	}

	var runtimeError plugin_entities.ErrorResponse
	if err := json.Unmarshal(received[0].Data, &runtimeError); err != nil {
		t.Fatalf("unmarshal runtime error: %v", err)
	}
	if runtimeError.ErrorType != "PluginRuntimeError" {
		t.Errorf("expected PluginRuntimeError, got %s", runtimeError.ErrorType)
	}
	if runtimeError.Message != wantMessage {
		t.Errorf("unexpected runtime error message: %s", runtimeError.Message)
	}
	if runtimeError.Args["request_id"] != "lambda-request-id" {
		t.Errorf("expected Lambda request ID, got %#v", runtimeError.Args["request_id"])
	}
	if runtimeError.Args["status_code"] != float64(wantStatusCode) {
		t.Errorf("expected status code %d, got %#v", wantStatusCode, runtimeError.Args["status_code"])
	}
}

func TestInvokeServerlessWithRetry_BodyClosedOnRetry(t *testing.T) {
	attemptCount := atomic.Int32{}
	bodiesClosed := atomic.Int32{}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		attempt := attemptCount.Add(1)
		if attempt < 2 {
			w.WriteHeader(http.StatusBadGateway)
			w.Write([]byte("bad gateway"))
		} else {
			w.WriteHeader(http.StatusOK)
			w.Write([]byte("success"))
		}
	}))
	defer server.Close()

	originalClient := server.Client()
	trackingTransport := &trackingRoundTripper{
		base:         originalClient.Transport,
		bodiesClosed: &bodiesClosed,
	}

	trackingClient := &http.Client{
		Transport: trackingTransport,
	}

	runtime := &ServerlessPluginRuntime{
		Client:                    trackingClient,
		MaxRetryTimes:             2,
		PluginMaxExecutionTimeout: 10,
	}

	response, err := runtime.invokeServerlessWithRetry(
		context.Background(),
		server.URL,
		"test-session",
		[]byte("test-data"),
		"",
	)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}
	response.Body.Close()

	if attemptCount.Load() != 2 {
		t.Errorf("Expected 2 attempts, got %d", attemptCount.Load())
	}

	if bodiesClosed.Load() < 1 {
		t.Errorf("Expected at least 1 body to be closed during retry, got %d", bodiesClosed.Load())
	}
}

type trackingRoundTripper struct {
	base         http.RoundTripper
	bodiesClosed *atomic.Int32
}

func (t *trackingRoundTripper) RoundTrip(req *http.Request) (*http.Response, error) {
	if t.base == nil {
		t.base = http.DefaultTransport
	}

	resp, err := t.base.RoundTrip(req)
	if err != nil {
		return resp, err
	}

	originalBody := resp.Body
	resp.Body = &trackingReadCloser{
		ReadCloser:   originalBody,
		bodiesClosed: t.bodiesClosed,
	}

	return resp, err
}

type trackingReadCloser struct {
	io.ReadCloser
	bodiesClosed *atomic.Int32
}

func (t *trackingReadCloser) Close() error {
	t.bodiesClosed.Add(1)
	return t.ReadCloser.Close()
}

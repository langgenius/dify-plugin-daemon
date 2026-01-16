package serverless_runtime

import (
	"io"
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"testing"
	"time"

	"github.com/langgenius/dify-plugin-daemon/pkg/entities"
	"github.com/langgenius/dify-plugin-daemon/pkg/entities/plugin_entities"
	"github.com/langgenius/dify-plugin-daemon/pkg/utils/mapping"
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

	response, err := runtime.invokeServerlessWithRetry(server.URL, "test-session", []byte("test-data"))
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
	response, err := runtime.invokeServerlessWithRetry(server.URL, "test-session", []byte("test-data"))
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

	response, err := runtime.invokeServerlessWithRetry(server.URL, "test-session", []byte("test-data"))
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

	response, err := runtime.invokeServerlessWithRetry(server.URL, "test-session", []byte("test-data"))
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

	response, err := runtime.invokeServerlessWithRetry(server.URL, "test-session", []byte("test-data"))

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

	_, err := runtime.invokeServerlessWithRetry(server.URL, "test-session", []byte("test-data"))
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

	response, err := runtime.invokeServerlessWithRetry(server.URL, "test-session", []byte("test-data"))
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
	_, err := runtime.invokeServerlessWithRetry(server.URL, "test-session-123", testData)
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

	response, err := runtime.invokeServerlessWithRetry(server.URL, "test-session", []byte("test-data"))
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

package transaction

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http/httptest"
	"sync/atomic"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/langgenius/dify-plugin-daemon/internal/core/io_tunnel/backwards_invocation"
	"github.com/langgenius/dify-plugin-daemon/pkg/entities/plugin_entities"
	"github.com/langgenius/dify-plugin-daemon/pkg/utils/parser"
)

func TestHandle_SessionNotFound_WritesErrorResponse(t *testing.T) {
	gin.SetMode(gin.TestMode)

	// Arrange: craft a valid PluginUniversalEvent that will trigger the session handler path
	// with a backwards_request_id extracted from SessionMessage.Data
	const (
		testSessionID       = "test-session-not-exist"
		testBackwardsReqID  = "req-123"
	)

	invokePayload := map[string]any{
		"backwards_request_id": testBackwardsReqID,
	}
	invokePayloadBytes := parser.MarshalJsonBytes(invokePayload)

	sessionMsg := plugin_entities.SessionMessage{
		Type: plugin_entities.SESSION_MESSAGE_TYPE_INVOKE,
		Data: invokePayloadBytes,
	}
	sessionMsgBytes := parser.MarshalJsonBytes(sessionMsg)

	event := plugin_entities.PluginUniversalEvent{
		SessionId: testSessionID,
		Event:     plugin_entities.PLUGIN_EVENT_SESSION,
		Data:      json.RawMessage(sessionMsgBytes),
	}
	body := bytes.NewReader(parser.MarshalJsonBytes(event))

	recorder := httptest.NewRecorder()
	ctx, _ := gin.CreateTestContext(recorder)
	ctx.Request = httptest.NewRequest("POST", "/serverless", body)

	h := NewServerlessTransactionHandler(2 * time.Second)

	// Act: handle the request; since no session exists in memory and the cache client
	// is not initialized in tests, session_manager.GetSession will fail and the
	// error branch should write a BackwardsInvocationResponseEvent to the response.
	h.Handle(ctx, "ignored")

	// Assert
	if recorder.Code != 400 {
		t.Fatalf("expected status code 400, got %d", recorder.Code)
	}

	var resp backwards_invocation.BackwardsInvocationResponseEvent
	if err := json.Unmarshal(recorder.Body.Bytes(), &resp); err != nil {
		t.Fatalf("response should be JSON BackwardsInvocationResponseEvent, got error: %v, body: %s", err, recorder.Body.String())
	}

	if resp.Event != backwards_invocation.REQUEST_EVENT_ERROR {
		t.Fatalf("expected event=error, got %q", resp.Event)
	}
	if resp.Message != "failed to get session info from cache" {
		t.Fatalf("expected message 'failed to get session info from cache', got %q", resp.Message)
	}
	if resp.BackwardsRequestId != testBackwardsReqID {
		t.Fatalf("expected backwards_request_id=%q, got %q", testBackwardsReqID, resp.BackwardsRequestId)
	}

	// Data should include error_type, detail, and session_id
	m, ok := resp.Data.(map[string]any)
	if !ok {
		// json.Unmarshal into interface{} yields map[string]any by default; if not, re-marshal and unmarshal to map
		raw := parser.MarshalJsonBytes(resp.Data)
		var tmp map[string]any
		if err := json.Unmarshal(raw, &tmp); err == nil {
			m = tmp
		} else {
			t.Fatalf("response data should be an object: %v", err)
		}
	}

	if v, _ := m["error_type"].(string); v != "SessionNotFound" {
		t.Fatalf("expected data.error_type=SessionNotFound, got %v", m["error_type"])
	}
	if v, _ := m["session_id"].(string); v != testSessionID {
		t.Fatalf("expected data.session_id=%q, got %v", testSessionID, m["session_id"]) 
	}
	if v, _ := m["detail"].(string); v == "" {
		t.Fatalf("expected non-empty data.detail, got empty")
	}
}

func TestServerlessTransactionWriteCloser_WriteRecoversFromPanic(t *testing.T) {
	w := &serverlessTransactionWriteCloser{
		done: make(chan bool),
		writer: func([]byte) (int, error) {
			panic("boom")
		},
		flush: func() {},
	}

	n, err := w.Write([]byte("payload"))
	if err == nil {
		t.Fatal("expected write error after panic")
	}
	if n != 0 {
		t.Fatalf("expected zero bytes written, got %d", n)
	}
	if got := err.Error(); got != "serverless transaction write panic: boom" {
		t.Fatalf("unexpected error: %q", got)
	}
	if atomic.LoadInt32(&w.closed) == 0 {
		t.Fatal("expected writer to be closed after panic")
	}
}

func TestServerlessTransactionWriteCloser_FlushRecoversFromPanic(t *testing.T) {
	w := &serverlessTransactionWriteCloser{
		done:   make(chan bool),
		writer: func(data []byte) (int, error) { return len(data), nil },
		flush: func() {
			panic("boom")
		},
	}

	w.Flush()

	if atomic.LoadInt32(&w.closed) == 0 {
		t.Fatal("expected writer to be closed after flush panic")
	}
}

func TestServerlessTransactionWriteCloser_WriteAfterClose(t *testing.T) {
	w := &serverlessTransactionWriteCloser{
		done:   make(chan bool),
		writer: func(data []byte) (int, error) { return len(data), nil },
		flush:  func() {},
	}

	if err := w.Close(); err != nil {
		t.Fatalf("close failed: %v", err)
	}

	_, err := w.Write([]byte("payload"))
	if err == nil {
		t.Fatal("expected closed pipe error")
	}
	if err != io.ErrClosedPipe {
		t.Fatalf("expected io.ErrClosedPipe, got %v", err)
	}
}

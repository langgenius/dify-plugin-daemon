package service

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/langgenius/dify-plugin-daemon/pkg/entities/model_entities"
	"github.com/langgenius/dify-plugin-daemon/pkg/utils/routine"
	"github.com/langgenius/dify-plugin-daemon/pkg/utils/stream"
	"github.com/stretchr/testify/assert"
)

func TestGetPluginMetricLabels(t *testing.T) {
	t.Run("nil session", func(t *testing.T) {
		pluginID, runtimeType := getPluginMetricLabels(nil)
		assert.Equal(t, "unknown", pluginID)
		assert.Equal(t, "unknown", runtimeType)
	})

	t.Run("session with nil runtime", func(t *testing.T) {
		// This test would require creating a mock session
		// For now, we just verify the function handles nil gracefully
		pluginID, runtimeType := getPluginMetricLabels(nil)
		assert.Equal(t, "unknown", pluginID)
		assert.Equal(t, "unknown", runtimeType)
	})
}

type fakeSSEWriter struct {
	lock       sync.Mutex
	header     http.Header
	body       bytes.Buffer
	status     int
	disconnect chan bool
}

func newFakeSSEWriter() *fakeSSEWriter {
	return &fakeSSEWriter{
		header:     http.Header{},
		disconnect: make(chan bool, 1),
	}
}

func (w *fakeSSEWriter) Header() http.Header {
	return w.header
}

func (w *fakeSSEWriter) Write(data []byte) (int, error) {
	w.lock.Lock()
	defer w.lock.Unlock()
	return w.body.Write(data)
}

func (w *fakeSSEWriter) WriteHeader(status int) {
	w.lock.Lock()
	defer w.lock.Unlock()
	w.status = status
}

func (w *fakeSSEWriter) Flush() {}

func (w *fakeSSEWriter) CloseNotify() <-chan bool {
	return w.disconnect
}

func (w *fakeSSEWriter) Body() string {
	w.lock.Lock()
	defer w.lock.Unlock()
	return w.body.String()
}

func llmChunk(content string) model_entities.LLMResultChunk {
	index := 0
	return model_entities.LLMResultChunk{
		Model: model_entities.LLMModel("test-model"),
		Delta: model_entities.LLMResultChunkDelta{
			Index: &index,
			Message: model_entities.PromptMessage{
				Role:    model_entities.PROMPT_MESSAGE_ROLE_ASSISTANT,
				Content: content,
			},
		},
	}
}

func runLLMSSE(
	t *testing.T,
	source *stream.Stream[model_entities.LLMResultChunk],
	maxTimeoutSeconds int,
	options sseOptions,
) (*fakeSSEWriter, string) {
	t.Helper()
	routine.InitPool(8)
	gin.SetMode(gin.TestMode)

	writer := newFakeSSEWriter()
	ctx, _ := gin.CreateTestContext(writer)
	ctx.Request = httptest.NewRequest(http.MethodPost, "/", nil)

	statuses := make(chan string, 4)
	baseSSEService(
		func() (*stream.Stream[model_entities.LLMResultChunk], error) {
			return source, nil
		},
		ctx,
		maxTimeoutSeconds,
		func(status string, duration float64) {
			statuses <- status
		},
		options,
	)

	select {
	case status := <-statuses:
		return writer, status
	case <-time.After(5 * time.Second):
		t.Fatal("baseSSEService never reported a completion status")
		return writer, ""
	}
}

func TestFirstTokenGrace(t *testing.T) {
	assert.Equal(t, 250*time.Millisecond, firstTokenGrace(time.Second))
	assert.Equal(t, 250*time.Millisecond, firstTokenGrace(5*time.Second))
	assert.Equal(t, time.Second, firstTokenGrace(20*time.Second))
}

func TestBaseSSEServiceStopsAnInvocationThatNeverProducesAToken(t *testing.T) {
	source := stream.NewStream[model_entities.LLMResultChunk](8)

	writer, status := runLLMSSE(t, source, 60, sseOptions{firstTokenBudget: 40 * time.Millisecond})

	assert.Equal(t, "first_token_timeout", status)
	assert.Contains(t, writer.Body(), FirstTokenTimeoutErrorType)
	assert.Contains(t, writer.Body(), "enforced_by")
	assert.True(t, source.IsClosed())
}

func TestBaseSSEServiceDoesNotCountAnEnvelopeFrameAsTheFirstToken(t *testing.T) {
	source := stream.NewStream[model_entities.LLMResultChunk](8)
	source.WriteBlocking(llmChunk(""))

	_, status := runLLMSSE(t, source, 60, sseOptions{firstTokenBudget: 40 * time.Millisecond})

	assert.Equal(t, "first_token_timeout", status)
}

func TestBaseSSEServiceOnlyBoundsTheWaitForTheFirstToken(t *testing.T) {
	source := stream.NewStream[model_entities.LLMResultChunk](8)
	source.WriteBlocking(llmChunk("hello"))

	observed := make(chan time.Duration, 1)
	go func() {
		time.Sleep(400 * time.Millisecond)
		source.Close()
	}()

	writer, status := runLLMSSE(t, source, 60, sseOptions{
		firstTokenBudget: 40 * time.Millisecond,
		onFirstToken:     func(elapsed time.Duration) { observed <- elapsed },
	})

	assert.Equal(t, "success", status)
	assert.NotContains(t, writer.Body(), FirstTokenTimeoutErrorType)
	assert.Less(t, <-observed, 290*time.Millisecond)
}

func TestBaseSSEServiceKeepsTheOverallTimeoutWhenItIsTighter(t *testing.T) {
	source := stream.NewStream[model_entities.LLMResultChunk](8)

	writer, status := runLLMSSE(t, source, 0, sseOptions{firstTokenBudget: 40 * time.Millisecond})

	assert.Equal(t, "timeout", status)
	assert.Contains(t, writer.Body(), "killed by timeout")
	assert.NotContains(t, writer.Body(), FirstTokenTimeoutErrorType)
}

func TestBaseSSEServiceReportsCompletionOnce(t *testing.T) {
	source := stream.NewStream[model_entities.LLMResultChunk](8)
	routine.InitPool(8)
	gin.SetMode(gin.TestMode)

	writer := newFakeSSEWriter()
	ctx, _ := gin.CreateTestContext(writer)
	ctx.Request = httptest.NewRequest(http.MethodPost, "/", nil)

	statuses := make(chan string, 8)
	baseSSEService(
		func() (*stream.Stream[model_entities.LLMResultChunk], error) {
			return source, nil
		},
		ctx,
		60,
		func(status string, duration float64) { statuses <- status },
		sseOptions{firstTokenBudget: 40 * time.Millisecond},
	)

	assert.Equal(t, "first_token_timeout", <-statuses)

	select {
	case extra := <-statuses:
		t.Fatalf("completion was reported twice, second status was %q", extra)
	case <-time.After(200 * time.Millisecond):
	}
}

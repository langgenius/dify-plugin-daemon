package calldify

import (
	"encoding/binary"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/langgenius/dify-plugin-daemon/internal/core/dify_invocation"
	"github.com/langgenius/dify-plugin-daemon/pkg/utils/http_requests"
	"github.com/langgenius/dify-plugin-daemon/pkg/utils/routine"
	"github.com/stretchr/testify/require"
)

func TestBackwardsLLMTimeoutInheritance(t *testing.T) {
	for _, tc := range []struct {
		name               string
		first, idle, total int64
		want               http_requests.StreamTimeouts
	}{
		{"legacy_defaults_remain_bounded", 0, 0, 0, http_requests.StreamTimeouts{FirstResponse: 240 * time.Second, ReadIdle: 240 * time.Second, Total: 240 * time.Second}},
		{"total_override_also_extends_first_response", 0, 0, 600000, http_requests.StreamTimeouts{FirstResponse: 600 * time.Second, ReadIdle: 240 * time.Second, Total: 600 * time.Second}},
		{"independent_budgets", 300000, 120000, 600000, http_requests.StreamTimeouts{FirstResponse: 300 * time.Second, ReadIdle: 120 * time.Second, Total: 600 * time.Second}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			invocation, err := NewDifyInvocationDaemon(NewDifyInvocationDaemonPayload{
				BaseUrl: "http://localhost", ReadTimeout: 240000,
				LLMFirstResponseTimeout: tc.first, LLMIdleTimeout: tc.idle, LLMTotalTimeout: tc.total,
			})
			require.NoError(t, err)
			require.Equal(t, tc.want, invocation.(*RealBackwardsInvocation).llmStreamTimeouts)
		})
	}
}

func TestBackwardsLLMEndpointsUseNewBudgets(t *testing.T) {
	routine.InitPool(16)
	for _, structured := range []bool{false, true} {
		for _, streaming := range []bool{false, true} {
			t.Run(fmt.Sprintf("structured=%v/stream=%v", structured, streaming), func(t *testing.T) {
				server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					require.Equal(t, "test-key", r.Header.Get("X-Inner-Api-Key"))
					path := "/inner/api/invoke/llm"
					if structured {
						path += "/structured-output"
					}
					require.Equal(t, path, r.URL.Path)
					w.WriteHeader(http.StatusOK)
					w.(http.Flusher).Flush()
					select {
					case <-r.Context().Done():
						return
					case <-time.After(140 * time.Millisecond):
					}
					body := []byte(`{"data":{"model":"fixture","delta":{"index":0,"message":{"role":"assistant","content":"ok"}},"structured_output":{"answer":"ok"}}}`)
					header := make([]byte, 14)
					header[0] = 0x0f
					binary.LittleEndian.PutUint16(header[2:4], 10)
					binary.LittleEndian.PutUint32(header[4:8], uint32(len(body)))
					_, _ = w.Write(append(header, body...))
				}))
				defer server.Close()
				invocation, err := NewDifyInvocationDaemon(NewDifyInvocationDaemonPayload{
					BaseUrl: server.URL, CallingKey: "test-key", ReadTimeout: 40, ResponseMaxBufferSize: 1024 * 1024,
					LLMTotalTimeout: 1000, LLMIdleTimeout: 60,
				})
				require.NoError(t, err)
				schema := dify_invocation.InvokeLLMSchema{Stream: streaming}
				if structured {
					response, err := invocation.InvokeLLMWithStructuredOutput(&dify_invocation.InvokeLLMWithStructuredOutputRequest{InvokeLLMSchema: schema})
					require.NoError(t, err)
					defer response.Close()
					require.True(t, response.Next())
					chunk, err := response.Read()
					require.NoError(t, err)
					require.Equal(t, "ok", chunk.Delta.Message.Content)
					require.Equal(t, "ok", chunk.StructuredOutput["answer"])
					require.False(t, response.Next())
				} else {
					response, err := invocation.InvokeLLM(&dify_invocation.InvokeLLMRequest{InvokeLLMSchema: schema})
					require.NoError(t, err)
					defer response.Close()
					require.True(t, response.Next())
					chunk, err := response.Read()
					require.NoError(t, err)
					require.Equal(t, "ok", chunk.Delta.Message.Content)
					require.False(t, response.Next())
				}
			})
		}
	}
}

func TestBackwardsToolRetainsLegacyTimeout(t *testing.T) {
	routine.InitPool(16)
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.(http.Flusher).Flush()
		<-r.Context().Done()
	}))
	defer server.Close()
	invocation, err := NewDifyInvocationDaemon(NewDifyInvocationDaemonPayload{
		BaseUrl: server.URL, ReadTimeout: 60, ResponseMaxBufferSize: 1024,
		LLMTotalTimeout: 1000,
	})
	require.NoError(t, err)
	response, err := invocation.InvokeTool(&dify_invocation.InvokeToolRequest{})
	require.NoError(t, err)
	defer response.Close()
	require.True(t, response.Next())
	_, err = response.Read()
	require.ErrorContains(t, err, "failed to read system header")
	var timeout *http_requests.StreamTimeoutError
	require.NotErrorAs(t, err, &timeout)
}

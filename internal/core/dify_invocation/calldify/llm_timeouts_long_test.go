package calldify

import (
	"encoding/binary"
	"net/http"
	"net/http/httptest"
	"os"
	"strconv"
	"testing"
	"time"

	"github.com/langgenius/dify-plugin-daemon/internal/core/dify_invocation"
	"github.com/langgenius/dify-plugin-daemon/pkg/entities/model_entities"
	"github.com/langgenius/dify-plugin-daemon/pkg/utils/http_requests"
	"github.com/langgenius/dify-plugin-daemon/pkg/utils/routine"
	"github.com/langgenius/dify-plugin-daemon/pkg/utils/stream"
	"github.com/stretchr/testify/require"
)

// Opt-in wall-clock regression against deployment budgets without provider
// credentials or external model calls. The two paths run concurrently.
func TestBackwardsLLMLongStream(t *testing.T) {
	secondsText := os.Getenv("DIFY_TEST_BACKWARDS_LLM_STREAM_SECONDS")
	if secondsText == "" {
		t.Skip("set DIFY_TEST_BACKWARDS_LLM_STREAM_SECONDS to exercise deployment budgets")
	}
	seconds, err := strconv.Atoi(secondsText)
	require.NoError(t, err)
	require.Greater(t, seconds, 0)
	require.LessOrEqual(t, seconds, 1200)
	readMS := func(key string, fallback int64) int64 {
		if value := os.Getenv(key); value != "" {
			parsed, err := strconv.ParseInt(value, 10, 64)
			require.NoError(t, err)
			return parsed
		}
		return fallback
	}
	payload := NewDifyInvocationDaemonPayload{
		ReadTimeout:             readMS("DIFY_BACKWARDS_INVOCATION_READ_TIMEOUT", 240000),
		LLMFirstResponseTimeout: readMS("DIFY_BACKWARDS_INVOCATION_LLM_FIRST_RESPONSE_TIMEOUT", 0),
		LLMIdleTimeout:          readMS("DIFY_BACKWARDS_INVOCATION_LLM_IDLE_TIMEOUT", 0),
		LLMTotalTimeout:         readMS("DIFY_BACKWARDS_INVOCATION_LLM_TOTAL_TIMEOUT", 0),
		ResponseMaxBufferSize:   1024 * 1024,
	}
	require.Greater(t, int64(seconds)*1000, payload.ReadTimeout, "stream must outlive the legacy body deadline")
	routine.InitPool(16)
	for _, legacy := range []bool{false, true} {
		name := "llm_policy"
		if legacy {
			name = "legacy_control"
		}
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusOK)
				w.(http.Flusher).Flush()
				body := []byte(`{"data":{"model":"fixture","delta":{"index":0,"message":{"role":"assistant","content":"progress"}}}}`)
				header := make([]byte, 14)
				header[0] = 0x0f
				binary.LittleEndian.PutUint16(header[2:4], 10)
				binary.LittleEndian.PutUint32(header[4:8], uint32(len(body)))
				frame := append(header, body...)
				ticker := time.NewTicker(time.Second)
				defer ticker.Stop()
				for range seconds {
					_, _ = w.Write(frame)
					w.(http.Flusher).Flush()
					select {
					case <-r.Context().Done():
						return
					case <-ticker.C:
					}
				}
			}))
			defer server.Close()
			config := payload
			config.BaseUrl = server.URL
			invocation, err := NewDifyInvocationDaemon(config)
			require.NoError(t, err)
			started := time.Now()
			var response *stream.Stream[model_entities.LLMResultChunk]
			if legacy {
				response, err = StreamResponse[model_entities.LLMResultChunk](invocation.(*RealBackwardsInvocation), http.MethodPost, "invoke/llm")
			} else {
				response, err = invocation.InvokeLLM(&dify_invocation.InvokeLLMRequest{InvokeLLMSchema: dify_invocation.InvokeLLMSchema{Stream: true}})
			}
			require.NoError(t, err)
			defer response.Close()
			count := 0
			for response.Next() {
				_, err = response.Read()
				if err != nil {
					break
				}
				count++
			}
			elapsed := time.Since(started)
			t.Logf("frames=%d elapsed=%s error=%v", count, elapsed, err)
			if legacy {
				// The timer may interrupt the frame prefix, header, or payload.
				// Assert the fixed cutoff rather than one platform-specific read error.
				require.Error(t, err)
				var timeout *http_requests.StreamTimeoutError
				require.NotErrorAs(t, err, &timeout)
				require.Greater(t, count, 0)
				require.Less(t, count, seconds)
				require.GreaterOrEqual(t, elapsed, time.Duration(config.ReadTimeout)*time.Millisecond)
				require.Less(t, elapsed, time.Duration(config.ReadTimeout)*time.Millisecond+5*time.Second)
			} else {
				require.NoError(t, err)
				require.Equal(t, seconds, count)
			}
		})
	}
}

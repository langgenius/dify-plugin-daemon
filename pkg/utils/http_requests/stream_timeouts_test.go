package http_requests

import (
	"context"
	"encoding/binary"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/langgenius/dify-plugin-daemon/pkg/utils/routine"
	"github.com/langgenius/dify-plugin-daemon/pkg/utils/stream"
	"github.com/stretchr/testify/require"
)

func timeoutTestFrame(body string) []byte {
	header := make([]byte, 14)
	header[0] = 0x0f
	binary.LittleEndian.PutUint16(header[2:4], 10)
	binary.LittleEndian.PutUint32(header[4:8], uint32(len(body)))
	return append(header, body...)
}

func readTimeoutTestStream(response *stream.Stream[map[string]int]) (int, error) {
	defer response.Close()
	count := 0
	for response.Next() {
		if _, err := response.Read(); err != nil {
			return count, err
		}
		count++
	}
	return count, nil
}

func requireTimeoutPhase(t *testing.T, err error, phase StreamTimeoutPhase) {
	t.Helper()
	var timeout *StreamTimeoutError
	require.ErrorAs(t, err, &timeout)
	require.Equal(t, phase, timeout.Phase)
	require.ErrorIs(t, err, context.DeadlineExceeded)
	require.True(t, timeout.Timeout())
	require.Contains(t, err.Error(), "phase="+string(phase))
	require.NotContains(t, err.Error(), "use of closed network connection")
}

func TestStreamTimeoutsContinuousProgressOutlivesLegacyBudget(t *testing.T) {
	routine.InitPool(16)
	for _, framed := range []bool{true, false} {
		t.Run(fmt.Sprintf("length_prefixed=%v", framed), func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusOK)
				w.(http.Flusher).Flush()
				for i := 0; i < 12; i++ {
					select {
					case <-r.Context().Done():
						return
					case <-time.After(20 * time.Millisecond):
					}
					body := fmt.Sprintf(`{"chunk":%d}`, i)
					if framed {
						_, _ = w.Write(timeoutTestFrame(body))
					} else {
						_, _ = fmt.Fprintf(w, "data: %s\n\n", body)
					}
					w.(http.Flusher).Flush()
				}
			}))
			defer server.Close()
			response, err := RequestAndParseStream[map[string]int](server.Client(), server.URL, http.MethodGet,
				HttpReadTimeout(50), HttpUsingLengthPrefixed(framed),
				HttpStreamTimeouts(StreamTimeouts{FirstResponse: time.Second, ReadIdle: 150 * time.Millisecond, Total: 2 * time.Second}))
			require.NoError(t, err)
			count, err := readTimeoutTestStream(response)
			require.NoError(t, err)
			require.Equal(t, 12, count)
		})
	}
}

func TestStreamTimeoutsClassifyAndCancelBlockedRequests(t *testing.T) {
	routine.InitPool(16)
	for _, tc := range []struct {
		name    string
		headers bool
		first   bool
		active  bool
		limits  StreamTimeouts
		phase   StreamTimeoutPhase
	}{
		{"before_headers", false, false, false, StreamTimeouts{80 * time.Millisecond, time.Second, 2 * time.Second}, StreamTimeoutFirstResponse},
		{"first_frame", true, false, false, StreamTimeouts{80 * time.Millisecond, time.Second, 2 * time.Second}, StreamTimeoutFirstResponse},
		{"read_idle", true, true, false, StreamTimeouts{time.Second, 80 * time.Millisecond, 2 * time.Second}, StreamTimeoutReadIdle},
		{"total_with_progress", true, true, true, StreamTimeouts{time.Second, time.Second, 160 * time.Millisecond}, StreamTimeoutTotal},
	} {
		t.Run(tc.name, func(t *testing.T) {
			cancelled := make(chan struct{})
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				defer close(cancelled)
				if tc.headers {
					w.WriteHeader(http.StatusOK)
					w.(http.Flusher).Flush()
				}
				if tc.first {
					_, _ = w.Write(timeoutTestFrame(`{"chunk":1}`))
					w.(http.Flusher).Flush()
				}
				if tc.active {
					for {
						select {
						case <-r.Context().Done():
							return
						case <-time.After(15 * time.Millisecond):
							_, _ = w.Write(timeoutTestFrame(`{"chunk":2}`))
							w.(http.Flusher).Flush()
						}
					}
				}
				<-r.Context().Done()
			}))
			defer server.Close()
			response, err := RequestAndParseStream[map[string]int](server.Client(), server.URL, http.MethodGet,
				HttpUsingLengthPrefixed(true), HttpStreamTimeouts(tc.limits))
			if tc.headers {
				require.NoError(t, err)
				_, err = readTimeoutTestStream(response)
			}
			requireTimeoutPhase(t, err, tc.phase)
			select {
			case <-cancelled:
			case <-time.After(time.Second):
				t.Fatal("upstream request was not cancelled")
			}
		})
	}
}

func TestStreamTimeoutsBlockingResponseDoesNotUseIdleBudget(t *testing.T) {
	routine.InitPool(16)
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.(http.Flusher).Flush()
		select {
		case <-time.After(180 * time.Millisecond):
			_, _ = w.Write(timeoutTestFrame(`{"chunk":1}`))
		case <-r.Context().Done():
		}
	}))
	defer server.Close()
	response, err := RequestAndParseStream[map[string]int](server.Client(), server.URL, http.MethodPost,
		HttpUsingLengthPrefixed(true), HttpStreamTimeouts(StreamTimeouts{time.Second, 50 * time.Millisecond, 2 * time.Second}))
	require.NoError(t, err)
	count, err := readTimeoutTestStream(response)
	require.NoError(t, err)
	require.Equal(t, 1, count)
}

func TestStreamTimeoutsPartialFrameBytesResetIdle(t *testing.T) {
	routine.InitPool(16)
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write(timeoutTestFrame(`{"chunk":1}`))
		w.(http.Flusher).Flush()
		for _, b := range timeoutTestFrame(`{"chunk":2}`) {
			select {
			case <-r.Context().Done():
				return
			case <-time.After(10 * time.Millisecond):
			}
			_, _ = w.Write([]byte{b})
			w.(http.Flusher).Flush()
		}
	}))
	defer server.Close()
	response, err := RequestAndParseStream[map[string]int](server.Client(), server.URL, http.MethodGet,
		HttpUsingLengthPrefixed(true), HttpStreamTimeouts(StreamTimeouts{time.Second, 100 * time.Millisecond, 2 * time.Second}))
	require.NoError(t, err)
	count, err := readTimeoutTestStream(response)
	require.NoError(t, err)
	require.Equal(t, 2, count)
}

func TestStreamTimeoutsParentCancellationAndConsumerClose(t *testing.T) {
	routine.InitPool(16)
	for _, consumerClose := range []bool{false, true} {
		t.Run(fmt.Sprintf("consumer_close=%v", consumerClose), func(t *testing.T) {
			cancelled := make(chan struct{})
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				_, _ = w.Write(timeoutTestFrame(`{"chunk":1}`))
				w.(http.Flusher).Flush()
				<-r.Context().Done()
				close(cancelled)
			}))
			defer server.Close()
			ctx, cancel := context.WithCancelCause(context.Background())
			defer cancel(nil)
			response, err := RequestAndParseStream[map[string]int](server.Client(), server.URL, http.MethodGet,
				HttpContext(ctx), HttpUsingLengthPrefixed(true), HttpStreamTimeouts(StreamTimeouts{time.Second, time.Second, 2 * time.Second}))
			require.NoError(t, err)
			require.True(t, response.Next())
			_, err = response.Read()
			require.NoError(t, err)
			if consumerClose {
				response.Close()
			} else {
				cause := errors.New("workflow cancelled")
				cancel(cause)
				_, err = readTimeoutTestStream(response)
				require.ErrorIs(t, err, cause)
				var timeout *StreamTimeoutError
				require.False(t, errors.As(err, &timeout))
			}
			select {
			case <-cancelled:
			case <-time.After(time.Second):
				t.Fatal("upstream request was not cancelled")
			}
		})
	}
}

func TestStreamTimeoutsPreserveUpstreamAndFramingErrors(t *testing.T) {
	routine.InitPool(16)
	for _, status := range []int{http.StatusBadRequest, http.StatusOK} {
		t.Run(fmt.Sprintf("status=%d", status), func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
				w.WriteHeader(status)
				_, _ = w.Write([]byte("invalid provider request"))
			}))
			defer server.Close()
			response, err := RequestAndParseStream[map[string]int](server.Client(), server.URL, http.MethodGet,
				HttpUsingLengthPrefixed(true), HttpStreamTimeouts(StreamTimeouts{time.Second, time.Second, 2 * time.Second}))
			if status == http.StatusOK {
				require.NoError(t, err)
				_, err = readTimeoutTestStream(response)
				require.Contains(t, err.Error(), "magic number mismatch")
			} else {
				require.Contains(t, err.Error(), "status code: 400")
				require.Contains(t, err.Error(), "invalid provider request")
			}
			var timeout *StreamTimeoutError
			require.False(t, errors.As(err, &timeout))
		})
	}
}

func TestLegacyStreamKeepsFixedBodyDeadline(t *testing.T) {
	routine.InitPool(16)
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.(http.Flusher).Flush()
		<-r.Context().Done()
	}))
	defer server.Close()
	response, err := RequestAndParseStream[map[string]int](server.Client(), server.URL, http.MethodGet,
		HttpUsingLengthPrefixed(true), HttpReadTimeout(80))
	require.NoError(t, err)
	_, err = readTimeoutTestStream(response)
	require.Error(t, err)
	var timeout *StreamTimeoutError
	require.False(t, errors.As(err, &timeout))
	require.Contains(t, err.Error(), "failed to read system header")
}

func TestStreamWatchdogStopAndConcurrentProgress(t *testing.T) {
	w, err := newStreamWatchdog(context.Background(), StreamTimeouts{time.Second, time.Second, 2 * time.Second})
	require.NoError(t, err)
	var wg sync.WaitGroup
	for i := 0; i < 8; i++ {
		wg.Go(func() {
			for j := 0; j < 100; j++ {
				_ = w.progress(true)
			}
		})
	}
	wg.Wait()
	require.NoError(t, w.stop())
	w.expire()
	require.Nil(t, w.cause)
	require.ErrorIs(t, context.Cause(w.ctx), context.Canceled)
	for _, limits := range []StreamTimeouts{{}, {-time.Second, time.Second, time.Second}, {time.Second, 0, time.Second}} {
		_, err := newStreamWatchdog(context.Background(), limits)
		require.Error(t, err)
		require.True(t, strings.Contains(err.Error(), "must be positive"))
	}
}

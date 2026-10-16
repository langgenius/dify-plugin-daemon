package service

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/require"

	"github.com/langgenius/dify-plugin-daemon/pkg/utils/log"
	"github.com/langgenius/dify-plugin-daemon/pkg/utils/stream"
)

func endpointResponseServer(
	t *testing.T,
	http2 bool,
	response *stream.Stream[[]byte],
	status int,
	headers http.Header,
	timeout time.Duration,
) (*httptest.Server, <-chan struct{}) {
	t.Helper()
	done := make(chan struct{})
	router := gin.New()
	router.Use(log.RecoveryMiddleware())
	router.GET("/", func(ctx *gin.Context) {
		defer close(done)
		sessionCtx, cancel := endpointSessionRequestContext(ctx.Request.Context(), timeout)
		defer cancel()
		writeEndpointResponse(ctx, sessionCtx, status, headers, response)
	})
	server := httptest.NewUnstartedServer(router)
	server.EnableHTTP2 = http2
	server.StartTLS()
	server.Client().Timeout = 5 * time.Second
	t.Cleanup(server.Close)
	return server, done
}

func awaitEndpointResponse(t *testing.T, done <-chan struct{}) {
	t.Helper()
	select {
	case <-done:
	case <-time.After(5 * time.Second):
		t.Fatal("endpoint response did not finish")
	}
}

func TestEndpointResponseFailures(t *testing.T) {
	for _, http2 := range []bool{false, true} {
		for _, partial := range []bool{false, true} {
			for _, timeout := range []bool{false, true} {
				t.Run(fmt.Sprintf("http2=%t/partial=%t/timeout=%t", http2, partial, timeout), func(t *testing.T) {
					var logs bytes.Buffer
					previous := slog.Default()
					slog.SetDefault(slog.New(slog.NewJSONHandler(&logs, nil)))
					t.Cleanup(func() { slog.SetDefault(previous) })
					response := stream.NewStream[[]byte](2)
					t.Cleanup(response.Close)
					err := errors.New("cannot open /private/example with token=test-secret")
					if partial {
						require.NoError(t, response.Write([]byte("prefix")))
					} else if !timeout {
						response.CloseWithError(err)
					}
					headers := http.Header{"Content-Type": {"application/octet-stream"}}
					if !partial {
						headers.Set("Content-Length", "999")
						headers.Set("Content-Encoding", "gzip")
					}
					limit := 5 * time.Second
					if timeout {
						limit = 200 * time.Millisecond
					}
					server, done := endpointResponseServer(t, http2, response, http.StatusCreated, headers, limit)
					r, requestErr := server.Client().Get(server.URL)
					require.NoError(t, requestErr)
					defer r.Body.Close()
					if http2 {
						require.Equal(t, 2, r.ProtoMajor)
					} else {
						require.Equal(t, 1, r.ProtoMajor)
					}
					if partial {
						require.Equal(t, http.StatusCreated, r.StatusCode)
						prefix := make([]byte, 6)
						_, readErr := io.ReadFull(r.Body, prefix)
						require.NoError(t, readErr)
						require.Equal(t, "prefix", string(prefix))
						if !timeout {
							response.CloseWithError(err)
						}
						rest, readErr := io.ReadAll(r.Body)
						require.Error(t, readErr)
						require.Empty(t, rest)
					} else {
						require.Equal(t, http.StatusInternalServerError, r.StatusCode)
						body, readErr := io.ReadAll(r.Body)
						require.NoError(t, readErr)
						require.Contains(t, string(body), "Internal Server Error")
						require.NotContains(t, string(body), "test-secret")
						require.Contains(t, r.Header.Get("Content-Type"), "application/json")
						require.Empty(t, r.Header.Get("Content-Encoding"))
					}
					awaitEndpointResponse(t, done)
					require.True(t, response.IsClosed())
					if timeout {
						require.Contains(t, logs.String(), context.DeadlineExceeded.Error())
					} else {
						require.Contains(t, logs.String(), err.Error())
					}
					require.NotContains(t, fmt.Sprint(r.Header), "test-secret")
				})
			}
		}
	}
}

func TestEndpointResponsePreservesPluginResponses(t *testing.T) {
	for _, status := range []int{http.StatusCreated, http.StatusNoContent, http.StatusBadRequest, http.StatusInternalServerError} {
		t.Run(fmt.Sprint(status), func(t *testing.T) {
			response := stream.NewStream[[]byte](2)
			body := []byte{0, 255, 1, 128}
			if status == http.StatusNoContent {
				body = nil
			} else {
				require.NoError(t, response.Write(body[:2]))
				require.NoError(t, response.Write(body[2:]))
			}
			response.Close()
			headers := http.Header{"Set-Cookie": {"first=1", "second=2"}, "Content-Type": {"application/octet-stream"}}
			server, done := endpointResponseServer(t, false, response, status, headers, time.Second)
			r, err := server.Client().Get(server.URL)
			require.NoError(t, err)
			defer r.Body.Close()
			data, err := io.ReadAll(r.Body)
			require.NoError(t, err)
			require.Equal(t, status, r.StatusCode)
			require.Equal(t, string(body), string(data))
			require.Equal(t, headers.Values("Set-Cookie"), r.Header.Values("Set-Cookie"))
			awaitEndpointResponse(t, done)
		})
	}
}

func TestEndpointResponseKeepsSessionAfterDisconnect(t *testing.T) {
	for _, finish := range []bool{false, true} {
		t.Run(fmt.Sprintf("plugin finishes=%t", finish), func(t *testing.T) {
			response := stream.NewStream[[]byte](2)
			t.Cleanup(response.Close)
			require.NoError(t, response.Write([]byte("prefix")))
			done := make(chan struct{})
			httpContexts := make(chan context.Context, 1)
			sessionContexts := make(chan context.Context, 1)
			router := gin.New()
			router.Use(log.RecoveryMiddleware())
			router.GET("/", func(ctx *gin.Context) {
				defer close(done)
				sessionCtx, cancel := endpointSessionRequestContext(ctx.Request.Context(), 300*time.Millisecond)
				defer cancel()
				httpContexts <- ctx.Request.Context()
				sessionContexts <- sessionCtx
				writeEndpointResponse(ctx, sessionCtx, http.StatusOK, nil, response)
			})
			server := httptest.NewServer(router)
			defer server.Close()
			requestCtx, cancel := context.WithCancel(context.Background())
			defer cancel()
			request, err := http.NewRequestWithContext(requestCtx, http.MethodGet, server.URL, nil)
			require.NoError(t, err)
			r, err := server.Client().Do(request)
			require.NoError(t, err)
			defer r.Body.Close()
			prefix := make([]byte, 6)
			_, err = io.ReadFull(r.Body, prefix)
			require.NoError(t, err)
			cancel()
			httpCtx, sessionCtx := <-httpContexts, <-sessionContexts
			awaitEndpointResponse(t, httpCtx.Done())
			require.NoError(t, sessionCtx.Err())
			require.False(t, response.IsClosed())
			select {
			case <-done:
				t.Fatal("handler closed the plugin session on HTTP disconnect")
			default:
			}
			if finish {
				require.NoError(t, response.Write([]byte("backwards invocation completed")))
				response.Close()
			}
			awaitEndpointResponse(t, done)
			require.True(t, response.IsClosed())
			if !finish {
				require.ErrorIs(t, sessionCtx.Err(), context.DeadlineExceeded)
			}
		})
	}
}

func TestEndpointResponseDeadlineStopsBlockedWrite(t *testing.T) {
	response := stream.NewStream[[]byte](1)
	t.Cleanup(response.Close)
	require.NoError(t, response.Write(bytes.Repeat([]byte("x"), 16*1024*1024)))
	done := make(chan struct{})
	router := gin.New()
	router.Use(log.RecoveryMiddleware())
	router.GET("/", func(ctx *gin.Context) {
		defer close(done)
		sessionCtx, cancel := endpointSessionRequestContext(ctx.Request.Context(), 200*time.Millisecond)
		defer cancel()
		writeEndpointResponse(ctx, sessionCtx, http.StatusOK, nil, response)
	})
	server := httptest.NewServer(router)
	defer server.Close()
	connection, err := net.DialTimeout("tcp", server.Listener.Addr().String(), time.Second)
	require.NoError(t, err)
	defer connection.Close()
	_, err = io.WriteString(connection, "GET / HTTP/1.1\r\nHost: test\r\n\r\n")
	require.NoError(t, err)
	awaitEndpointResponse(t, done)
	require.True(t, response.IsClosed())
}

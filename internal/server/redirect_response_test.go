package server

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/require"

	"github.com/langgenius/dify-plugin-daemon/pkg/utils/log"
)

type failedResponseReader struct{ err error }

func (r failedResponseReader) Read([]byte) (int, error) { return 0, r.err }

type trackedResponseBody struct {
	io.Reader
	closed bool
}

func (r *trackedResponseBody) Close() error {
	r.closed = true
	return nil
}

func TestRedirectedResponseBeforeHeaders(t *testing.T) {
	var logs bytes.Buffer
	previous := slog.Default()
	slog.SetDefault(slog.New(slog.NewJSONHandler(&logs, nil)))
	t.Cleanup(func() { slog.SetDefault(previous) })
	err := errors.New("cannot open /private/example with token=test-secret")
	body := &trackedResponseBody{Reader: failedResponseReader{err}}
	router := gin.New()
	router.GET("/", func(ctx *gin.Context) {
		writeRedirectedResponse(ctx, http.StatusCreated, http.Header{
			"Content-Length": {"999"}, "Content-Encoding": {"gzip"}, "Content-Type": {"application/octet-stream"},
		}, body)
	})
	recorder := httptest.NewRecorder()
	router.ServeHTTP(recorder, httptest.NewRequest(http.MethodGet, "/", nil))
	require.Equal(t, http.StatusInternalServerError, recorder.Code)
	require.Contains(t, recorder.Body.String(), "Internal Server Error")
	require.NotContains(t, recorder.Body.String(), "test-secret")
	require.Empty(t, recorder.Header().Get("Content-Encoding"))
	require.Empty(t, recorder.Header().Get("Content-Length"))
	require.Contains(t, recorder.Header().Get("Content-Type"), "application/json")
	require.Contains(t, logs.String(), err.Error())
	require.True(t, body.closed)
}

func TestRedirectedResponsePropagatesAbortAcrossNodes(t *testing.T) {
	for _, http2 := range []bool{false, true} {
		t.Run(fmt.Sprintf("http2=%t", http2), func(t *testing.T) {
			source := &trackedResponseBody{Reader: io.MultiReader(strings.NewReader("prefix"), failedResponseReader{errors.New("/private/example")})}
			done := make(chan struct{})
			upstream := gin.New()
			upstream.Use(log.RecoveryMiddleware())
			upstream.GET("/", func(ctx *gin.Context) {
				defer close(done)
				writeRedirectedResponse(ctx, http.StatusCreated, http.Header{"Content-Type": {"application/octet-stream"}}, source)
			})
			first := httptest.NewUnstartedServer(upstream)
			first.EnableHTTP2 = http2
			first.StartTLS()
			defer first.Close()
			first.Client().Timeout = 5 * time.Second
			forwarder := gin.New()
			forwarder.Use(log.RecoveryMiddleware())
			forwarder.GET("/", func(ctx *gin.Context) {
				r, err := first.Client().Get(first.URL)
				if err != nil {
					t.Errorf("redirect request failed: %v", err)
					ctx.AbortWithStatus(http.StatusBadGateway)
					return
				}
				writeRedirectedResponse(ctx, r.StatusCode, r.Header, r.Body)
			})
			second := httptest.NewUnstartedServer(forwarder)
			second.EnableHTTP2 = http2
			second.StartTLS()
			defer second.Close()
			second.Client().Timeout = 5 * time.Second
			r, err := second.Client().Get(second.URL)
			require.NoError(t, err)
			defer r.Body.Close()
			require.Equal(t, http.StatusCreated, r.StatusCode)
			if http2 {
				require.Equal(t, 2, r.ProtoMajor)
			} else {
				require.Equal(t, 1, r.ProtoMajor)
			}
			data, err := io.ReadAll(r.Body)
			require.Error(t, err)
			require.Equal(t, "prefix", string(data))
			<-done
			require.True(t, source.closed)
		})
	}
}

func TestRedirectedResponsePreservesPluginResponses(t *testing.T) {
	for _, status := range []int{http.StatusNoContent, http.StatusCreated, http.StatusInternalServerError} {
		t.Run(fmt.Sprint(status), func(t *testing.T) {
			data := "\x00\xffplugin error"
			if status == http.StatusNoContent {
				data = ""
			}
			body := &trackedResponseBody{Reader: strings.NewReader(data)}
			headers := http.Header{"Set-Cookie": {"first=1", "second=2"}, "Content-Type": {"application/octet-stream"}}
			router := gin.New()
			router.GET("/", func(ctx *gin.Context) { writeRedirectedResponse(ctx, status, headers, body) })
			recorder := httptest.NewRecorder()
			router.ServeHTTP(recorder, httptest.NewRequest(http.MethodGet, "/", nil))
			require.Equal(t, status, recorder.Code)
			require.Equal(t, data, recorder.Body.String())
			require.Equal(t, headers.Values("Set-Cookie"), recorder.Header().Values("Set-Cookie"))
			require.True(t, body.closed)
		})
	}
}

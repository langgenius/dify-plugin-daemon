package log

import (
	"bytes"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/require"
)

func TestRecoveryMiddlewarePreservesHTTPAbort(t *testing.T) {
	for name, value := range map[string]any{"http abort": http.ErrAbortHandler, "ordinary panic": "cannot open /private/example"} {
		t.Run(name, func(t *testing.T) {
			var logs bytes.Buffer
			previous := slog.Default()
			slog.SetDefault(slog.New(slog.NewJSONHandler(&logs, nil)))
			t.Cleanup(func() { slog.SetDefault(previous) })
			router := gin.New()
			router.Use(RecoveryMiddleware())
			router.GET("/", func(*gin.Context) { panic(value) })
			recorder := httptest.NewRecorder()
			serve := func() { router.ServeHTTP(recorder, httptest.NewRequest(http.MethodGet, "/", nil)) }
			if value == http.ErrAbortHandler {
				require.PanicsWithValue(t, http.ErrAbortHandler, serve)
				require.Empty(t, logs.String())
			} else {
				require.NotPanics(t, serve)
				require.Equal(t, http.StatusInternalServerError, recorder.Code)
				require.Empty(t, recorder.Body.String())
				require.Contains(t, logs.String(), "/private/example")
				require.Contains(t, logs.String(), "stack_trace")
			}
		})
	}
}

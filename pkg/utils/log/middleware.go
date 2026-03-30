package log

import (
	"fmt"
	"log/slog"
	"net/http"
	"runtime"
	"time"

	"github.com/gin-gonic/gin"
)

func TraceMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx := c.Request.Context()

		traceparent := c.GetHeader("traceparent")
		traceID, spanID, ok := ParseTraceparent(traceparent)
		if !ok {
			traceID = GenerateTraceID()
			spanID = GenerateSpanID()
		}
		ctx = WithTrace(ctx, TraceContext{
			TraceID: traceID,
			SpanID:  spanID,
		})

		identity := Identity{
			TenantID: c.Param("tenant_id"),
			UserID:   c.GetHeader("X-User-ID"),
			UserType: c.GetHeader("X-User-Type"),
		}
		ctx = WithIdentity(ctx, identity)

		c.Request = c.Request.WithContext(ctx)
		c.Next()
	}
}

type LoggerConfig struct {
	SkipPaths []string
}

func LoggerMiddleware() gin.HandlerFunc {
	return LoggerMiddlewareWithConfig(LoggerConfig{})
}

func LoggerMiddlewareWithConfig(config LoggerConfig) gin.HandlerFunc {
	skipPaths := make(map[string]bool)
	for _, path := range config.SkipPaths {
		skipPaths[path] = true
	}

	return func(c *gin.Context) {
		if skipPaths[c.Request.URL.Path] {
			c.Next()
			return
		}

		start := time.Now()
		path := c.Request.URL.Path
		query := c.Request.URL.RawQuery

		c.Next()

		latency := time.Since(start)
		status := c.Writer.Status()
		method := c.Request.Method
		clientIP := c.ClientIP()

		if query != "" {
			path = path + "?" + query
		}

		ctx := c.Request.Context()
		level := slog.LevelInfo
		if status >= 500 {
			level = slog.LevelError
		} else if status >= 400 {
			level = slog.LevelWarn
		}

		slog.Log(ctx, level, "HTTP request",
			"method", method,
			"path", path,
			"status", status,
			"latency_ms", latency.Milliseconds(),
			"client_ip", clientIP,
		)
	}
}

func RecoveryMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		defer func() {
			if err := recover(); err != nil {
				stack := captureFullStack()

				ctx := c.Request.Context()
				slog.ErrorContext(ctx, "panic recovered",
					"error", fmt.Sprintf("%v", err),
					"stack_trace", stack,
				)

				c.AbortWithStatus(http.StatusInternalServerError)
			}
		}()
		c.Next()
	}
}

func captureFullStack() string {
	buf := make([]byte, 4096)
	for {
		n := runtime.Stack(buf, false)
		if n < len(buf) {
			return string(buf[:n])
		}
		buf = make([]byte, len(buf)*2)
	}
}

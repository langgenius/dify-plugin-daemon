package log

import (
	"context"
	"fmt"
	"io"
	"log/slog"
	"os"
	"path/filepath"
	"runtime"
	"time"

	"github.com/gin-gonic/gin"
)

const ServiceName = "dify-plugin-daemon"

func Init(json bool, filename string) (io.Closer, error) {
	var w io.Writer = os.Stdout
	var closer io.Closer
	if filename != "" {
		dir := filepath.Dir(filename)
		if dir != "." && dir != "" {
			if err := os.MkdirAll(dir, 0755); err != nil {
				return nil, fmt.Errorf("create log directory %q: %w", dir, err)
			}
		}
		file, err := os.OpenFile(filename, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
		if err != nil {
			return nil, fmt.Errorf("open log file %q: %w", filename, err)
		}
		w = io.MultiWriter(os.Stdout, file)
		closer = file
	}
	handler := NewHandler(Options{
		Level:   slog.LevelInfo,
		Service: ServiceName,
		JSON:    json,
		Out:     w,
	})
	slog.SetDefault(slog.New(handler))
	setupGinDebug()
	return closer, nil
}

func setupGinDebug() {
	gin.DebugPrintRouteFunc = func(httpMethod, absolutePath, handlerName string, nuHandlers int) {
		slog.Debug("gin route registered",
			"method", httpMethod,
			"path", absolutePath,
			"handler", handlerName,
			"handlers_count", nuHandlers,
		)
	}
}

func logWithCaller(ctx context.Context, level slog.Level, msg string, args ...any) {
	logger := slog.Default()
	if !logger.Enabled(ctx, level) {
		return
	}
	var pc uintptr
	var pcs [1]uintptr
	runtime.Callers(3, pcs[:])
	pc = pcs[0]
	r := slog.NewRecord(time.Now(), level, msg, pc)
	r.Add(args...)
	_ = logger.Handler().Handle(ctx, r)
}

func Debug(msg string, args ...any) {
	logWithCaller(context.Background(), slog.LevelDebug, msg, args...)
}

func Info(msg string, args ...any) {
	logWithCaller(context.Background(), slog.LevelInfo, msg, args...)
}

func Warn(msg string, args ...any) {
	logWithCaller(context.Background(), slog.LevelWarn, msg, args...)
}

func Error(msg string, args ...any) {
	logWithCaller(context.Background(), slog.LevelError, msg, args...)
}

func Panic(msg string, args ...any) {
	logWithCaller(context.Background(), slog.LevelError, msg, args...)
	panic(msg)
}

func DebugContext(ctx context.Context, msg string, args ...any) {
	logWithCaller(ctx, slog.LevelDebug, msg, args...)
}

func InfoContext(ctx context.Context, msg string, args ...any) {
	logWithCaller(ctx, slog.LevelInfo, msg, args...)
}

func WarnContext(ctx context.Context, msg string, args ...any) {
	logWithCaller(ctx, slog.LevelWarn, msg, args...)
}

func ErrorContext(ctx context.Context, msg string, args ...any) {
	logWithCaller(ctx, slog.LevelError, msg, args...)
}

func PanicContext(ctx context.Context, msg string, args ...any) {
	logWithCaller(ctx, slog.LevelError, msg, args...)
	panic(msg)
}

func RecoverAndExit() {
	if err := recover(); err != nil {
		stack := captureFullPanicStack()
		slog.Error("panic recovered",
			"error", fmt.Sprintf("%v", err),
			"stack_trace", stack,
		)
		os.Exit(1)
	}
}

func captureFullPanicStack() string {
	buf := make([]byte, 4096)
	for {
		n := runtime.Stack(buf, false)
		if n < len(buf) {
			return string(buf[:n])
		}
		buf = make([]byte, len(buf)*2)
	}
}

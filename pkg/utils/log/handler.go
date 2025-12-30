package log

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"runtime"
	"strconv"
	"strings"
	"sync"
	"time"
)

type Options struct {
	Level   slog.Leveler
	Service string
	JSON    bool
	Out     io.Writer
}

type Handler struct {
	opts   Options
	mu     *sync.Mutex
	attrs  []slog.Attr
	groups []string
}

func NewHandler(opts Options) *Handler {
	if opts.Level == nil {
		opts.Level = slog.LevelInfo
	}
	return &Handler{
		opts: opts,
		mu:   &sync.Mutex{},
	}
}

func (h *Handler) Enabled(_ context.Context, level slog.Level) bool {
	return level >= h.opts.Level.Level()
}

func (h *Handler) Handle(ctx context.Context, r slog.Record) error {
	fields := make(map[string]any)

	fields["ts"] = r.Time.UTC().Format(time.RFC3339Nano)
	fields["severity"] = levelToSeverity(r.Level)
	fields["service"] = h.opts.Service
	fields["caller"] = getCaller(r.PC)
	fields["message"] = r.Message

	if tc, ok := TraceFromContext(ctx); ok && tc.TraceID != "" {
		fields["trace_id"] = tc.TraceID
		if tc.SpanID != "" {
			fields["span_id"] = tc.SpanID
		}
	}

	if id, ok := IdentityFromContext(ctx); ok {
		identity := make(map[string]any)
		if id.TenantID != "" {
			identity["tenant_id"] = id.TenantID
		}
		if id.UserID != "" {
			identity["user_id"] = id.UserID
		}
		if id.UserType != "" {
			identity["user_type"] = id.UserType
		}
		if len(identity) > 0 {
			fields["identity"] = identity
		}
	}

	attrs := make(map[string]any)
	for _, attr := range h.attrs {
		collectAttr(attrs, attr)
	}
	r.Attrs(func(a slog.Attr) bool {
		collectAttr(attrs, a)
		return true
	})

	var stackTrace string
	if st, ok := attrs["stack_trace"]; ok {
		if s, isStr := st.(string); isStr {
			stackTrace = s
		}
		delete(attrs, "stack_trace")
	}

	if errVal, ok := attrs["error"]; ok {
		if err, isErr := errVal.(error); isErr {
			attrs["error"] = err.Error()
			if r.Level >= slog.LevelError && stackTrace == "" {
				stackTrace = captureStackTrace()
			}
		}
	}

	if len(attrs) > 0 {
		fields["attributes"] = attrs
	}

	if stackTrace != "" {
		fields["stack_trace"] = stackTrace
	}

	h.mu.Lock()
	defer h.mu.Unlock()

	if h.opts.JSON {
		return h.writeJSON(fields)
	}
	return h.writeText(fields)
}

func (h *Handler) WithAttrs(attrs []slog.Attr) slog.Handler {
	newAttrs := make([]slog.Attr, len(h.attrs)+len(attrs))
	copy(newAttrs, h.attrs)
	copy(newAttrs[len(h.attrs):], attrs)
	return &Handler{
		opts:   h.opts,
		mu:     h.mu,
		attrs:  newAttrs,
		groups: h.groups,
	}
}

func (h *Handler) WithGroup(name string) slog.Handler {
	if name == "" {
		return h
	}
	newGroups := make([]string, len(h.groups)+1)
	copy(newGroups, h.groups)
	newGroups[len(h.groups)] = name
	return &Handler{
		opts:   h.opts,
		mu:     h.mu,
		attrs:  h.attrs,
		groups: newGroups,
	}
}

func (h *Handler) writeJSON(fields map[string]any) error {
	data, err := json.Marshal(fields)
	if err != nil {
		return err
	}
	_, err = h.opts.Out.Write(append(data, '\n'))
	return err
}

func (h *Handler) writeText(fields map[string]any) error {
	var sb strings.Builder

	sb.WriteString(fields["ts"].(string))
	sb.WriteByte(' ')
	sb.WriteString(fields["severity"].(string))
	sb.WriteByte(' ')
	sb.WriteString(fields["service"].(string))
	sb.WriteByte(' ')
	sb.WriteString(fields["caller"].(string))

	if traceID, ok := fields["trace_id"].(string); ok {
		sb.WriteString(" trace_id=")
		sb.WriteString(traceID)
	}

	if identity, ok := fields["identity"].(map[string]any); ok {
		if tenantID, ok := identity["tenant_id"].(string); ok {
			sb.WriteString(" tenant_id=")
			sb.WriteString(tenantID)
		}
		if userID, ok := identity["user_id"].(string); ok {
			sb.WriteString(" user_id=")
			sb.WriteString(userID)
		}
	}

	sb.WriteByte(' ')
	sb.WriteString(fields["message"].(string))

	if attrs, ok := fields["attributes"].(map[string]any); ok {
		for k, v := range attrs {
			sb.WriteByte(' ')
			sb.WriteString(k)
			sb.WriteByte('=')
			sb.WriteString(formatValue(v))
		}
	}

	sb.WriteByte('\n')

	if stackTrace, ok := fields["stack_trace"].(string); ok && stackTrace != "" {
		sb.WriteString(stackTrace)
		sb.WriteByte('\n')
	}

	_, err := h.opts.Out.Write([]byte(sb.String()))
	return err
}

func levelToSeverity(level slog.Level) string {
	switch {
	case level >= slog.LevelError:
		return "ERROR"
	case level >= slog.LevelWarn:
		return "WARN"
	case level >= slog.LevelInfo:
		return "INFO"
	default:
		return "DEBUG"
	}
}

func getCaller(pc uintptr) string {
	if pc == 0 {
		return "unknown"
	}
	fs := runtime.CallersFrames([]uintptr{pc})
	f, _ := fs.Next()
	if f.File == "" {
		return "unknown"
	}
	idx := strings.LastIndex(f.File, "/")
	if idx >= 0 {
		f.File = f.File[idx+1:]
	}
	return f.File + ":" + strconv.Itoa(f.Line)
}

func collectAttr(attrs map[string]any, a slog.Attr) {
	if a.Key == "" {
		return
	}
	switch a.Value.Kind() {
	case slog.KindGroup:
		group := make(map[string]any)
		for _, ga := range a.Value.Group() {
			collectAttr(group, ga)
		}
		if len(group) > 0 {
			attrs[a.Key] = group
		}
	default:
		attrs[a.Key] = a.Value.Any()
	}
}

func formatValue(v any) string {
	switch val := v.(type) {
	case string:
		if strings.ContainsAny(val, " \t\n\"") {
			return strconv.Quote(val)
		}
		return val
	case error:
		return strconv.Quote(val.Error())
	default:
		return fmt.Sprintf("%v", v)
	}
}

func captureStackTrace() string {
	const depth = 32
	var pcs [depth]uintptr
	n := runtime.Callers(5, pcs[:])
	if n == 0 {
		return ""
	}

	var sb strings.Builder
	frames := runtime.CallersFrames(pcs[:n])
	for {
		frame, more := frames.Next()
		if strings.Contains(frame.File, "runtime/") {
			if !more {
				break
			}
			continue
		}
		sb.WriteString(frame.Function)
		sb.WriteByte('\n')
		sb.WriteByte('\t')
		sb.WriteString(frame.File)
		sb.WriteByte(':')
		sb.WriteString(strconv.Itoa(frame.Line))
		sb.WriteByte('\n')
		if !more {
			break
		}
	}
	return sb.String()
}

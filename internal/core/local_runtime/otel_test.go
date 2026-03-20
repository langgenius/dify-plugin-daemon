package local_runtime

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"
	gootel "go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/propagation"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	"go.opentelemetry.io/otel/sdk/trace/tracetest"
)

func TestPythonInitSpansReuseUpstreamTrace(t *testing.T) {
	sr := tracetest.NewSpanRecorder()
	tp := sdktrace.NewTracerProvider()
	tp.RegisterSpanProcessor(sr)
	gootel.SetTracerProvider(tp)
	gootel.SetTextMapPropagator(propagation.TraceContext{})

	ctx := context.Background()
	tr := gootel.Tracer("test")
	ctx, parent := tr.Start(ctx, "parent")

	runtime := &LocalPluginRuntime{}
	runtime.SetTraceContext(ctx)

	_, child := runtime.startSpan("python.init_env")
	child.End()
	parent.End()

	spans := sr.Ended()
	require.Len(t, spans, 2)

	parentSpan := spans[0]
	childSpan := spans[1]
	if parentSpan.Name() == "python.init_env" {
		parentSpan, childSpan = childSpan, parentSpan
	}
	require.Equal(t, parentSpan.SpanContext().TraceID(), childSpan.SpanContext().TraceID())
}

func TestStartSpanDoesNotAffectSubsequentSpans(t *testing.T) {
	sr := tracetest.NewSpanRecorder()
	tp := sdktrace.NewTracerProvider()
	tp.RegisterSpanProcessor(sr)
	gootel.SetTracerProvider(tp)
	gootel.SetTextMapPropagator(propagation.TraceContext{})

	ctx := context.Background()
	tr := gootel.Tracer("test")
	ctx, parent := tr.Start(ctx, "parent")

	runtime := &LocalPluginRuntime{}
	runtime.SetTraceContext(ctx)

	ctx1, span1 := runtime.startSpan("span1")
	span1.End()

	ctx2, span2 := runtime.startSpan("span2")

	require.False(t, ctx1.Err() != nil, "ctx1 should not be canceled yet")
	require.False(t, ctx2.Err() != nil, "ctx2 should not be canceled")

	span2.End()
	parent.End()

	spans := sr.Ended()
	require.Len(t, spans, 3)

	traceID := spans[0].SpanContext().TraceID()
	for _, span := range spans {
		require.Equal(t, traceID, span.SpanContext().TraceID(), "all spans should share the same trace ID")
	}
}

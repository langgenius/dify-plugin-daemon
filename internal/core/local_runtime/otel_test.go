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
	// setup tracer provider with span recorder
	sr := tracetest.NewSpanRecorder()
	tp := sdktrace.NewTracerProvider()
	tp.RegisterSpanProcessor(sr)
	gootel.SetTracerProvider(tp)
	gootel.SetTextMapPropagator(propagation.TraceContext{})

	// create parent span with known trace id
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

	// ensure child trace id matches parent trace id
	parentSpan := spans[0]
	childSpan := spans[1]
	if parentSpan.Name() == "python.init_env" {
		parentSpan, childSpan = childSpan, parentSpan
	}
	require.Equal(t, parentSpan.SpanContext().TraceID(), childSpan.SpanContext().TraceID())
}

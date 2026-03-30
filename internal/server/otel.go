package server

import (
	"context"
	"net/http"
	"time"

	"github.com/langgenius/dify-plugin-daemon/internal/types/app"
	"github.com/langgenius/dify-plugin-daemon/pkg/utils/log"

	otelhttp "go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
	gootel "go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/exporters/otlp/otlpmetric/otlpmetrichttp"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracehttp"
	"go.opentelemetry.io/otel/propagation"
	sdkmetric "go.opentelemetry.io/otel/sdk/metric"
	"go.opentelemetry.io/otel/sdk/resource"
	"go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.24.0"
)

// InitTelemetry sets up OTLP HTTP exporters for traces and metrics, ParentBased sampling to reuse upstream trace decisions,
// and global propagators for W3C TraceContext + Baggage. It returns a shutdown func.
func InitTelemetry(cfg *app.Config) (func(context.Context) error, error) {
	// Build resource
	res, err := resource.New(context.Background(),
		resource.WithAttributes(
			semconv.ServiceName("dify-plugin-daemon"),
			attribute.String("service.version", "unknown"),
		),
	)
	if err != nil {
		return nil, err
	}

	// --- Traces exporter (OTLP HTTP) ---
	traceEndpoint := cfg.OtlpTraceEndpoint
	if traceEndpoint == "" {
		traceEndpoint = cfg.OtlpBaseEndpoint + "/v1/traces"
	}

	traceClient := &http.Client{Timeout: 10 * time.Second}
	_ = traceClient // reserved for future customization
	traceExp, err := otlptracehttp.New(context.Background(),
		otlptracehttp.WithEndpointURL(traceEndpoint),
		otlptracehttp.WithHeaders(authHeaders(cfg.OtelApiKey)),
	)
	if err != nil {
		return nil, err
	}

	// Sampler: reuse parent decision, otherwise sample by ratio
	if cfg.OtelSamplingRate < 0 || cfg.OtelSamplingRate > 1 {
		cfg.OtelSamplingRate = 1.0
	}
	sam := trace.ParentBased(trace.TraceIDRatioBased(cfg.OtelSamplingRate))

	// Batch span processor options
	bsp := trace.NewBatchSpanProcessor(traceExp,
		trace.WithMaxQueueSize(cfg.OtelMaxQueueSize),
		trace.WithMaxExportBatchSize(cfg.OtelMaxExportBatchSize),
		trace.WithBatchTimeout(time.Duration(cfg.OtelBatchScheduleDelayMS)*time.Millisecond),
		trace.WithExportTimeout(time.Duration(cfg.OtelBatchExportTimeoutMS)*time.Millisecond),
	)

	tp := trace.NewTracerProvider(
		trace.WithSampler(sam),
		trace.WithResource(res),
		trace.WithSpanProcessor(tenantIDSpanProcessor{}),
		trace.WithSpanProcessor(bsp),
	)
	gootel.SetTracerProvider(tp)

	// Instrument default HTTP transport so all http.DefaultClient and clients using default RoundTripper get propagation and spans
	http.DefaultTransport = otelhttp.NewTransport(http.DefaultTransport)

	// --- Metrics exporter (OTLP HTTP) ---
	metricEndpoint := cfg.OtlpMetricEndpoint
	if metricEndpoint == "" {
		metricEndpoint = cfg.OtlpBaseEndpoint + "/v1/metrics"
	}

	metricExp, err := otlpmetrichttp.New(context.Background(),
		otlpmetrichttp.WithEndpointURL(metricEndpoint),
		otlpmetrichttp.WithHeaders(authHeaders(cfg.OtelApiKey)),
	)
	if err != nil {
		log.Warn("otel metric exporter init failed", "error", err)
	} else {
		prd := sdkmetric.NewPeriodicReader(metricExp,
			sdkmetric.WithInterval(time.Duration(cfg.OtelMetricExportIntervalMS)*time.Millisecond),
			sdkmetric.WithTimeout(time.Duration(cfg.OtelMetricExportTimeoutMS)*time.Millisecond),
		)
		mp := sdkmetric.NewMeterProvider(
			sdkmetric.WithReader(prd),
			sdkmetric.WithResource(res),
		)
		gootel.SetMeterProvider(mp)
	}

	// Propagators: W3C TraceContext + Baggage
	gootel.SetTextMapPropagator(propagation.NewCompositeTextMapPropagator(
		propagation.TraceContext{}, propagation.Baggage{},
	))

	return func(ctx context.Context) error {
		ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
		defer cancel()
		if err := tp.Shutdown(ctx); err != nil {
			return err
		}
		if mp := gootel.GetMeterProvider(); mp != nil {
			if mp := gootel.GetMeterProvider(); mp != nil {
				if s, ok := mp.(interface{ Shutdown(context.Context) error }); ok && s != nil {
					_ = s.Shutdown(ctx)
				}
			}
		}
		return nil
	}, nil
}

func authHeaders(apiKey string) map[string]string {
	if apiKey == "" {
		return nil
	}
	return map[string]string{
		"Authorization": "Bearer " + apiKey,
	}
}

// tenantIDSpanProcessor injects tenant_id attribute into spans when present in context.
type tenantIDSpanProcessor struct{}

func (tenantIDSpanProcessor) OnStart(ctx context.Context, s trace.ReadWriteSpan) {
	if id, ok := log.IdentityFromContext(ctx); ok && id.TenantID != "" {
		s.SetAttributes(attribute.String("tenant_id", id.TenantID))
	}
}

func (tenantIDSpanProcessor) OnEnd(trace.ReadOnlySpan) {}

func (tenantIDSpanProcessor) Shutdown(context.Context) error { return nil }

func (tenantIDSpanProcessor) ForceFlush(context.Context) error { return nil }

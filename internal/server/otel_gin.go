package server

import (
	"github.com/gin-gonic/gin"
	otelgin "go.opentelemetry.io/contrib/instrumentation/github.com/gin-gonic/gin/otelgin"
	gootel "go.opentelemetry.io/otel"
)

// OtelGinMiddleware returns a Gin middleware that instruments requests using the global tracer provider
// and extracts upstream trace context (traceparent/baggage) automatically.
func OtelGinMiddleware() gin.HandlerFunc {
	return otelgin.Middleware("dify-plugin-daemon",
		otelgin.WithTracerProvider(gootel.GetTracerProvider()), otelgin.WithPropagators(gootel.GetTextMapPropagator()))
}

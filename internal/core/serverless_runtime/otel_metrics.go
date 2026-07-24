package serverless_runtime

import (
	"context"
	"sync"

	"github.com/langgenius/dify-plugin-daemon/internal/core/io_tunnel/access_types"
	"github.com/langgenius/dify-plugin-daemon/pkg/utils/log"
	gootel "go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/metric"
)

const (
	serverlessRuntimeMetricScope        = "dify-plugin-daemon/serverless-runtime"
	serverlessRuntimeFailuresMetricName = "plugin.serverless.runtime.failures"
)

var (
	serverlessRuntimeFailuresOnce    sync.Once
	serverlessRuntimeFailuresCounter metric.Int64Counter
	serverlessRuntimeFailuresErr     error
)

func recordServerlessRuntimeFailure(
	action access_types.PluginAccessAction,
	statusCode int,
	failureType string,
) {
	serverlessRuntimeFailuresOnce.Do(func() {
		serverlessRuntimeFailuresCounter, serverlessRuntimeFailuresErr = gootel.Meter(
			serverlessRuntimeMetricScope,
		).Int64Counter(
			serverlessRuntimeFailuresMetricName,
			metric.WithDescription("Number of failed requests to a serverless plugin Function URL."),
			metric.WithUnit("{failure}"),
		)
		if serverlessRuntimeFailuresErr != nil {
			log.Warn("failed to init serverless runtime failure counter", "error", serverlessRuntimeFailuresErr)
		}
	})
	if serverlessRuntimeFailuresErr != nil {
		return
	}

	serverlessRuntimeFailuresCounter.Add(
		context.Background(),
		1,
		metric.WithAttributes(
			attribute.String("plugin.action", string(action)),
			attribute.Int("http.response.status_code", statusCode),
			attribute.String("failure.type", failureType),
			attribute.String("failure.source", "lambda_function_url"),
		),
	)
}

func serverlessRuntimeFailureType(statusCode int) string {
	switch statusCode {
	case 0:
		return "transport"
	case 429:
		return "rate_limit"
	case 502:
		return "gateway"
	default:
		return "http"
	}
}

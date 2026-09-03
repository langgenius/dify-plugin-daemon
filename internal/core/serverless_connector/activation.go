package serverless

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"time"

	"github.com/langgenius/dify-plugin-daemon/pkg/utils/http_requests"
	"github.com/langgenius/dify-plugin-daemon/pkg/utils/parser"
)

type activationRequest struct {
	InstanceID string `json:"instance_id"`
}

type activationResponse struct {
	Ready    bool   `json:"ready"`
	Endpoint string `json:"endpoint,omitempty"`
}

// ErrActivationTimeout indicates the plugin was not woken up and ready before the
// activation deadline elapsed. Callers should treat it as a failed invocation.
var ErrActivationTimeout = fmt.Errorf("timed out waiting for plugin to become ready")

// Activate performs the activation preflight against the serverless connector.
// It wakes a scaled-to-zero plugin (renewing its activity lease) and blocks until
// the plugin reports ready. instanceID is the connector function name of the
// target plugin. timeout bounds how long the daemon waits for readiness; the
// connector applies its own backstop as well.
//
// It returns nil once the plugin is ready, ErrActivationTimeout when the plugin
// is not ready within the window, or a wrapped error for any other failure.
func Activate(ctx context.Context, instanceID string, timeout time.Duration) error {
	activateURL, err := url.JoinPath(baseurl.String(), "/v1/activation/activate")
	if err != nil {
		return err
	}

	if timeout > 0 {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, timeout)
		defer cancel()
	}

	response, err := http_requests.Request(
		client,
		activateURL,
		"POST",
		http_requests.HttpContext(ctx),
		http_requests.HttpHeader(map[string]string{
			"Authorization": SERVERLESS_CONNECTOR_API_KEY,
		}),
		http_requests.HttpPayloadJson(activationRequest{InstanceID: instanceID}),
	)
	if err != nil {
		if ctx.Err() != nil {
			return ErrActivationTimeout
		}
		return fmt.Errorf("failed to request serverless connector activation: %w", err)
	}
	defer response.Body.Close()

	body, _ := io.ReadAll(io.LimitReader(response.Body, 4*1024))

	switch response.StatusCode {
	case http.StatusOK:
		parsed, err := parser.UnmarshalJsonBytes[activationResponse](body)
		if err != nil {
			return fmt.Errorf("failed to parse activation response: %w", err)
		}
		if !parsed.Ready {
			return ErrActivationTimeout
		}
		return nil
	case http.StatusGatewayTimeout:
		return ErrActivationTimeout
	default:
		return fmt.Errorf("unexpected response from serverless connector activation: status=%d body=%s", response.StatusCode, string(body))
	}
}

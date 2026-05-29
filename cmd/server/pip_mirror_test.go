package main

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/langgenius/dify-plugin-daemon/internal/types/app"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestParseCloudflareTraceLocation(t *testing.T) {
	loc, err := parseCloudflareTraceLocation(strings.NewReader("fl=29f\n h=cloudflare.com\nloc = cn\n malformed\n"))
	require.NoError(t, err)
	assert.Equal(t, "CN", loc)
}

func TestParseCloudflareTraceLocationMissingLoc(t *testing.T) {
	_, err := parseCloudflareTraceLocation(strings.NewReader("fl=29f\ncolo=SJC\n"))
	require.Error(t, err)
	assert.Contains(t, err.Error(), "missing loc")
}

func TestDetectAndApplyPipMirror(t *testing.T) {
	tests := []struct {
		name             string
		config           app.Config
		statusCode       int
		body             string
		expectedMirror   string
		expectedLoc      string
		expectedApplied  bool
		expectedRequests int
		expectErr        string
	}{
		{
			name: "disabled skips detection",
			config: app.Config{
				PipMirrorAutoDetect: false,
			},
			expectedRequests: 0,
		},
		{
			name: "configured mirror is preserved",
			config: app.Config{
				PipMirrorAutoDetect: true,
				PipMirrorUrl:        "https://mirror.example/simple",
			},
			expectedMirror:   "https://mirror.example/simple",
			expectedRequests: 0,
		},
		{
			name: "cn applies alibaba mirror",
			config: app.Config{
				PipMirrorAutoDetect: true,
			},
			statusCode:       http.StatusOK,
			body:             "ip=1.1.1.1\nloc=CN\n",
			expectedMirror:   alibabaCloudPypiMirrorURL,
			expectedLoc:      "CN",
			expectedApplied:  true,
			expectedRequests: 1,
		},
		{
			name: "non-cn leaves mirror empty",
			config: app.Config{
				PipMirrorAutoDetect: true,
			},
			statusCode:       http.StatusOK,
			body:             "ip=1.1.1.1\nloc=US\n",
			expectedLoc:      "US",
			expectedRequests: 1,
		},
		{
			name: "invalid trace returns error",
			config: app.Config{
				PipMirrorAutoDetect: true,
			},
			statusCode:       http.StatusOK,
			body:             "ip=1.1.1.1\ncolo=SJC\n",
			expectedRequests: 1,
			expectErr:        "missing loc",
		},
		{
			name: "unexpected status returns error",
			config: app.Config{
				PipMirrorAutoDetect: true,
			},
			statusCode:       http.StatusBadGateway,
			expectedRequests: 1,
			expectErr:        "status 502 Bad Gateway",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			requests := 0
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				requests++
				w.WriteHeader(tt.statusCode)
				_, _ = w.Write([]byte(tt.body))
			}))
			defer server.Close()

			loc, applied, err := detectAndApplyPipMirror(&tt.config, server.Client(), server.URL)

			if tt.expectErr != "" {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.expectErr)
			} else {
				require.NoError(t, err)
			}

			assert.Equal(t, tt.expectedLoc, loc)
			assert.Equal(t, tt.expectedApplied, applied)
			assert.Equal(t, tt.expectedMirror, tt.config.PipMirrorUrl)
			assert.Equal(t, tt.expectedRequests, requests)
		})
	}
}

func TestApplyPipMirrorAutoDetectDoesNotBlockStartupOnDetectionFailure(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusBadGateway)
	}))
	defer server.Close()

	originalTraceURL := cloudflareTraceURL
	cloudflareTraceURL = server.URL
	defer func() {
		cloudflareTraceURL = originalTraceURL
	}()

	config := app.Config{
		PipMirrorAutoDetect: true,
	}

	applyPipMirrorAutoDetect(&config)

	assert.Empty(t, config.PipMirrorUrl)
}

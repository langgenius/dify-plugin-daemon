package main

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/langgenius/dify-plugin-daemon/internal/types/app"
	"github.com/stretchr/testify/assert"
)

func newInstantServer(statusCode int) *httptest.Server {
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(statusCode)
	}))
}

func newSlowServer(delay time.Duration, statusCode int) *httptest.Server {
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(delay)
		w.WriteHeader(statusCode)
	}))
}

func TestDetectAndApplyPipMirrorDisabled(t *testing.T) {
	config := app.Config{PipMirrorAutoDetect: false}
	mirror := detectAndApplyPipMirror(&config, &http.Client{Timeout: time.Second}, nil, "http://unused")
	assert.Empty(t, mirror)
	assert.Empty(t, config.PipMirrorUrl)
}

func TestDetectAndApplyPipMirrorExplicitURLPreserved(t *testing.T) {
	config := app.Config{
		PipMirrorAutoDetect: true,
		PipMirrorUrl:        "https://mirror.example/simple",
	}
	mirror := detectAndApplyPipMirror(&config, &http.Client{Timeout: time.Second}, nil, "http://unused")
	assert.Empty(t, mirror)
	assert.Equal(t, "https://mirror.example/simple", config.PipMirrorUrl)
}

func TestDetectAndApplyPipMirrorCandidateFasterThanOfficial(t *testing.T) {
	official := newSlowServer(80*time.Millisecond, http.StatusOK)
	defer official.Close()
	candidate := newInstantServer(http.StatusOK)
	defer candidate.Close()

	config := app.Config{PipMirrorAutoDetect: true}
	mirror := detectAndApplyPipMirror(&config, &http.Client{Timeout: 2 * time.Second}, []string{candidate.URL}, official.URL)
	assert.Equal(t, candidate.URL, mirror)
	assert.Equal(t, candidate.URL, config.PipMirrorUrl)
}

func TestDetectAndApplyPipMirrorOfficialFasterNoCandidateSelected(t *testing.T) {
	official := newInstantServer(http.StatusOK)
	defer official.Close()
	candidate := newSlowServer(80*time.Millisecond, http.StatusOK)
	defer candidate.Close()

	config := app.Config{PipMirrorAutoDetect: true}
	mirror := detectAndApplyPipMirror(&config, &http.Client{Timeout: 2 * time.Second}, []string{candidate.URL}, official.URL)
	assert.Empty(t, mirror)
	assert.Empty(t, config.PipMirrorUrl)
}

func TestDetectAndApplyPipMirrorOfficialUnreachableCandidateSelected(t *testing.T) {
	candidate := newInstantServer(http.StatusOK)
	defer candidate.Close()

	config := app.Config{PipMirrorAutoDetect: true}
	mirror := detectAndApplyPipMirror(
		&config,
		&http.Client{Timeout: 200 * time.Millisecond},
		[]string{candidate.URL},
		"http://127.0.0.1:1",
	)
	assert.Equal(t, candidate.URL, mirror)
	assert.Equal(t, candidate.URL, config.PipMirrorUrl)
}

func TestDetectAndApplyPipMirrorPicksFastestCandidate(t *testing.T) {
	official := newSlowServer(120*time.Millisecond, http.StatusOK)
	defer official.Close()
	slow := newSlowServer(80*time.Millisecond, http.StatusOK)
	defer slow.Close()
	fast := newInstantServer(http.StatusOK)
	defer fast.Close()

	config := app.Config{PipMirrorAutoDetect: true}
	mirror := detectAndApplyPipMirror(&config, &http.Client{Timeout: 2 * time.Second}, []string{slow.URL, fast.URL}, official.URL)
	assert.Equal(t, fast.URL, mirror)
}

func TestDetectAndApplyPipMirrorAllUnreachableNoMirrorSet(t *testing.T) {
	config := app.Config{PipMirrorAutoDetect: true}
	mirror := detectAndApplyPipMirror(
		&config,
		&http.Client{Timeout: 200 * time.Millisecond},
		[]string{"http://127.0.0.1:1"},
		"http://127.0.0.1:2",
	)
	assert.Empty(t, mirror)
	assert.Empty(t, config.PipMirrorUrl)
}

func TestDetectAndApplyPipMirrorCandidateErrorStatusIgnored(t *testing.T) {
	official := newInstantServer(http.StatusOK)
	defer official.Close()
	candidate := newInstantServer(http.StatusServiceUnavailable)
	defer candidate.Close()

	config := app.Config{PipMirrorAutoDetect: true}
	mirror := detectAndApplyPipMirror(&config, &http.Client{Timeout: time.Second}, []string{candidate.URL}, official.URL)
	assert.Empty(t, mirror)
	assert.Empty(t, config.PipMirrorUrl)
}

func TestSelectFastestMirrorEmptyCandidates(t *testing.T) {
	official := newInstantServer(http.StatusOK)
	defer official.Close()

	ctx := context.Background()
	result := selectFastestMirror(ctx, &http.Client{Timeout: time.Second}, nil, official.URL)
	assert.Empty(t, result)
}

func TestApplyPipMirrorAutoDetectDoesNotBlockStartupOnAllFailures(t *testing.T) {
	config := app.Config{
		PipMirrorAutoDetect: true,
		PipMirrorCandidates: []string{"http://127.0.0.1:1"},
	}
	applyPipMirrorAutoDetect(&config)
	assert.Empty(t, config.PipMirrorUrl)
}

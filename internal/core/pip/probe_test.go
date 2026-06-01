package pip

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestProbeMirrorReachable(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("index"))
	}))
	defer server.Close()

	status := probeMirror(context.Background(), server.Client(), Mirror{Name: "test", URL: server.URL})
	assert.True(t, status.Reachable)
	assert.Equal(t, http.StatusOK, status.StatusCode)
	assert.Equal(t, server.URL, status.URL)
	assert.Equal(t, "test", status.Name)
	assert.Empty(t, status.Error)
	assert.False(t, status.CheckedAt.IsZero())
}

func TestProbeMirrorUnexpectedStatus(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusBadGateway)
	}))
	defer server.Close()

	status := probeMirror(context.Background(), server.Client(), Mirror{URL: server.URL})
	assert.False(t, status.Reachable)
	assert.Equal(t, http.StatusBadGateway, status.StatusCode)
	assert.Contains(t, status.Error, "502")
}

func TestProbeMirrorConnectionError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {}))
	url := server.URL
	server.Close() // close immediately so the request fails

	status := probeMirror(context.Background(), &http.Client{Timeout: time.Second}, Mirror{URL: url})
	assert.False(t, status.Reachable)
	assert.NotEmpty(t, status.Error)
}

func TestNewProberDefaults(t *testing.T) {
	prober := NewProber(NewConfigProvider("", nil), 0, 0)
	assert.Equal(t, DefaultProbeTimeout, prober.client.Timeout)
	assert.Equal(t, DefaultProbeInterval, prober.interval)
}

func TestNewProberClampsIntervalToMinimum(t *testing.T) {
	prober := NewProber(NewConfigProvider("", nil), time.Second, 3*time.Second)
	assert.Equal(t, MinProbeInterval, prober.interval)
}

func TestProbeOnceMarksSelectedCandidate(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	official := server.URL + "/official"
	selected := server.URL + "/mirror"
	provider := NewCompositeProvider(
		selected, // configured mirror -> effective
		official,
		[]Mirror{{Name: "pypi", URL: official}, {Name: "mirror", URL: selected}},
		nil,
	)

	prober := NewProber(provider, time.Second, time.Hour)
	result := prober.ProbeOnce(context.Background())

	require.Len(t, result.Mirrors, 2)
	assert.Equal(t, selected, result.Selected)
	selectedStatus, ok := result.SelectedStatus()
	require.True(t, ok)
	assert.Equal(t, selected, selectedStatus.URL)
}

func TestProberStartStoresResult(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	provider := NewCompositeProvider("", server.URL, []Mirror{{Name: "pypi", URL: server.URL}}, nil)
	prober := NewProber(provider, time.Second, time.Hour)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	go prober.Start(ctx)

	require.Eventually(t, func() bool {
		result, ok := GetResult()
		if !ok {
			return false
		}
		status, found := result.SelectedStatus()
		return found && status.Reachable
	}, 2*time.Second, 20*time.Millisecond)
}

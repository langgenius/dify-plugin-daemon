package pip

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"sync/atomic"
	"time"
)

const (
	// DefaultPyPIIndexURL is the official PyPI simple index used for the
	// connectivity probe.
	DefaultPyPIIndexURL = "https://pypi.org/simple/"
	// DefaultProbeTimeout is the default per-probe timeout.
	DefaultProbeTimeout = 5 * time.Second
	// DefaultProbeInterval is the default interval between background probes.
	DefaultProbeInterval = 60 * time.Second
	// MinProbeInterval is the smallest allowed interval between background probes.
	// It guards against hammering the index with an overly aggressive interval.
	MinProbeInterval = 10 * time.Second
)

// MirrorStatus describes the most recent connectivity probe to a single mirror.
type MirrorStatus struct {
	Name       string    `json:"name"`
	URL        string    `json:"url"`
	Reachable  bool      `json:"reachable"`
	StatusCode int       `json:"status_code,omitempty"`
	LatencyMs  int64     `json:"latency_ms"`
	Error      string    `json:"error,omitempty"`
	Selected   bool      `json:"selected"`
	CheckedAt  time.Time `json:"checked_at"`
}

// ProbeResult is a snapshot of the latest probe cycle across all candidates. It
// is purely informational (surfaced on the health check / admin endpoints) and
// never affects whether the daemon is considered healthy.
type ProbeResult struct {
	// Selected is the effective mirror URL at the time of the probe ("" was
	// normalized to the official index URL).
	Selected  string         `json:"selected"`
	Mirrors   []MirrorStatus `json:"mirrors"`
	CheckedAt time.Time      `json:"checked_at"`
}

// SelectedStatus returns the status of the currently selected mirror, if probed.
func (r ProbeResult) SelectedStatus() (MirrorStatus, bool) {
	for _, m := range r.Mirrors {
		if m.Selected {
			return m, true
		}
	}
	return MirrorStatus{}, false
}

// currentResult holds the latest ProbeResult produced by a Prober. It is read by
// the health check / admin handlers and written by the background prober.
var currentResult atomic.Value // stores ProbeResult

// SetResult stores the latest probe result.
func SetResult(r ProbeResult) {
	currentResult.Store(r)
}

// GetResult returns the latest probe result. The boolean is false when no probe
// cycle has completed yet (or probing is disabled).
func GetResult() (ProbeResult, bool) {
	v := currentResult.Load()
	if v == nil {
		return ProbeResult{}, false
	}
	return v.(ProbeResult), true
}

// probeMirror performs a single connectivity check against a mirror and returns
// its status. It never returns an error; failures are captured inside the status.
func probeMirror(ctx context.Context, client *http.Client, mirror Mirror) MirrorStatus {
	start := time.Now()
	status := MirrorStatus{Name: mirror.Name, URL: mirror.URL, CheckedAt: start}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, mirror.URL, nil)
	if err != nil {
		status.Error = err.Error()
		status.LatencyMs = time.Since(start).Milliseconds()
		return status
	}

	resp, err := client.Do(req)
	if err != nil {
		status.Error = err.Error()
		status.LatencyMs = time.Since(start).Milliseconds()
		return status
	}
	defer resp.Body.Close()
	// drain a small amount to allow connection reuse; body content is irrelevant
	_, _ = io.Copy(io.Discard, io.LimitReader(resp.Body, 1024))

	status.StatusCode = resp.StatusCode
	status.LatencyMs = time.Since(start).Milliseconds()
	status.Reachable = resp.StatusCode == http.StatusOK
	if !status.Reachable {
		status.Error = fmt.Sprintf("unexpected status %s", resp.Status)
	}
	return status
}

// Prober periodically probes all candidate mirrors exposed by a PypiProvider and
// caches the latest ProbeResult.
type Prober struct {
	provider PypiProvider
	client   *http.Client
	interval time.Duration
}

// NewProber builds a Prober. Zero/empty arguments fall back to the package
// defaults; the interval is clamped to MinProbeInterval.
func NewProber(provider PypiProvider, timeout, interval time.Duration) *Prober {
	if timeout <= 0 {
		timeout = DefaultProbeTimeout
	}
	if interval <= 0 {
		interval = DefaultProbeInterval
	}
	if interval < MinProbeInterval {
		interval = MinProbeInterval
	}
	return &Prober{
		provider: provider,
		client:   &http.Client{Timeout: timeout},
		interval: interval,
	}
}

// ProbeOnce probes every candidate mirror once and returns the aggregated result.
func (p *Prober) ProbeOnce(ctx context.Context) ProbeResult {
	candidates := p.provider.Candidates()
	effective := p.provider.MirrorURL()
	official := officialURLOf(p.provider)

	selected := effective
	if selected == "" {
		selected = official
	}

	statuses := make([]MirrorStatus, 0, len(candidates))
	for _, candidate := range candidates {
		status := probeMirror(ctx, p.client, candidate)
		status.Selected = isSelected(candidate.URL, effective, official)
		statuses = append(statuses, status)
	}

	return ProbeResult{
		Selected:  selected,
		Mirrors:   statuses,
		CheckedAt: time.Now(),
	}
}

// Start runs an immediate probe cycle, stores the result, then keeps probing on
// the configured interval until ctx is cancelled. It is intended to run in its
// own goroutine for the lifetime of the daemon.
func (p *Prober) Start(ctx context.Context) {
	SetResult(p.ProbeOnce(ctx))

	ticker := time.NewTicker(p.interval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			SetResult(p.ProbeOnce(ctx))
		}
	}
}

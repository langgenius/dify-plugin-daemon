package pip

import (
	"context"
	"sync/atomic"
	"time"

	"github.com/langgenius/dify-plugin-daemon/internal/types/app"
	"github.com/langgenius/dify-plugin-daemon/pkg/utils/log"
)

// sharedProvider holds the process-wide PypiProvider. It is wrapped in a holder
// struct so that atomic.Value always stores the same concrete type.
var sharedProvider atomic.Value // stores providerHolder

type providerHolder struct{ provider PypiProvider }

// SetProvider installs the process-wide PypiProvider.
func SetProvider(p PypiProvider) {
	sharedProvider.Store(providerHolder{provider: p})
}

// Provider returns the process-wide PypiProvider, or nil if not initialized.
func Provider() PypiProvider {
	v := sharedProvider.Load()
	if v == nil {
		return nil
	}
	return v.(providerHolder).provider
}

// ProviderOrConfig returns the shared provider when initialized, otherwise a
// read-only ConfigProvider built from config. Used by runtimes that may be
// constructed before/without Bootstrap (e.g. in tests).
func ProviderOrConfig(config *app.Config) PypiProvider {
	if p := Provider(); p != nil {
		return p
	}
	return NewConfigProvider(config.PipMirrorUrl, configCandidates(config))
}

// probeEnabled reports whether the background PyPI probe should run. The probe
// only makes sense for the local runtime; in serverless mode the daemon never
// installs dependencies against PyPI directly.
func probeEnabled(config *app.Config) bool {
	return config.PipPypiProbeEnabled && config.Platform == app.PLATFORM_LOCAL
}

// officialURL resolves the official PyPI index URL from config.
func officialURL(config *app.Config) string {
	return DefaultPyPIIndexURL
}

// configCandidates builds the candidate list from config: the official index
// first, followed by the configured "name=url" candidates.
func configCandidates(config *app.Config) []Mirror {
	candidates := make([]Mirror, 0, 4)
	candidates = append(candidates, Mirror{Name: "pypi", URL: officialURL(config)})
	candidates = append(candidates, ParseCandidates(config.PipMirrorCandidates)...)
	return candidates
}

// NewProviderFromConfig builds a CompositeProvider from config, backed by the
// given store (which may be nil to disable DB resolution/selection).
func NewProviderFromConfig(config *app.Config, store MirrorStore) *CompositeProvider {
	return NewCompositeProvider(
		config.PipMirrorUrl,
		officialURL(config),
		configCandidates(config),
		store,
	)
}

// Bootstrap installs the process-wide CompositeProvider (database-backed) and,
// for the local runtime with probing enabled, starts the background prober.
//
// It must be called after the database has been initialized. The probe only
// runs for the local runtime: in serverless mode dependency installation is
// delegated to the serverless connector, so the daemon never talks to PyPI.
func Bootstrap(config *app.Config) {
	provider := NewProviderFromConfig(config, newDBStore())
	SetProvider(provider)

	if !probeEnabled(config) {
		return
	}

	prober := NewProber(
		provider,
		time.Duration(config.PipPypiProbeTimeout)*time.Second,
		time.Duration(config.PipPypiProbeInterval)*time.Second,
	)
	log.Info("starting PyPI connectivity probe", "candidates", len(provider.Candidates()))
	go prober.Start(context.Background())
}

// MirrorListItem describes a candidate mirror together with its latest probe
// measurement (if any).
type MirrorListItem struct {
	Mirror
	Selected   bool       `json:"selected"`
	Reachable  bool       `json:"reachable"`
	LatencyMs  int64      `json:"latency_ms"`
	StatusCode int        `json:"status_code,omitempty"`
	Error      string     `json:"error,omitempty"`
	CheckedAt  *time.Time `json:"checked_at,omitempty"`
}

// MirrorListing is the response shape for the "list candidates + latency" API.
type MirrorListing struct {
	Selected string           `json:"selected"`
	Mirrors  []MirrorListItem `json:"mirrors"`
}

// BuildMirrorListing combines a provider's candidates with the latest cached
// probe measurements into a listing suitable for the admin API.
func BuildMirrorListing(provider PypiProvider) MirrorListing {
	candidates := provider.Candidates()
	effective := provider.MirrorURL()
	official := officialURLOf(provider)

	selected := effective
	if selected == "" {
		selected = official
	}

	measured := make(map[string]MirrorStatus)
	if result, ok := GetResult(); ok {
		for _, s := range result.Mirrors {
			measured[s.URL] = s
		}
	}

	items := make([]MirrorListItem, 0, len(candidates))
	for _, candidate := range candidates {
		item := MirrorListItem{
			Mirror:   candidate,
			Selected: isSelected(candidate.URL, effective, official),
		}
		if status, ok := measured[candidate.URL]; ok {
			item.Reachable = status.Reachable
			item.LatencyMs = status.LatencyMs
			item.StatusCode = status.StatusCode
			item.Error = status.Error
			checkedAt := status.CheckedAt
			item.CheckedAt = &checkedAt
		}
		items = append(items, item)
	}

	return MirrorListing{Selected: selected, Mirrors: items}
}

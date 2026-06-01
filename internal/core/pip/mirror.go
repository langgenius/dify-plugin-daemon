package pip

import (
	"errors"
	"strings"
)

// Mirror is a candidate PyPI index (the value passed to pip/uv via -i).
type Mirror struct {
	Name string `json:"name"`
	URL  string `json:"url"`
}

// PypiProvider resolves the effective pip mirror and exposes the candidate
// mirrors that can be probed and selected.
//
// It is intentionally an interface so the resolution strategy can evolve: today
// it resolves from the database or static config, but later implementations may
// pick a mirror dynamically (e.g. the fastest reachable one).
type PypiProvider interface {
	// MirrorURL returns the effective pip index URL. An empty string means "use
	// the default/official PyPI index" (i.e. do not pass -i to pip/uv).
	MirrorURL() string
	// Candidates returns the candidate mirrors available for probing/selection.
	Candidates() []Mirror
}

// MutableProvider is a PypiProvider that additionally supports selecting and
// persisting a mirror (e.g. backed by the database).
type MutableProvider interface {
	PypiProvider
	// Select persists the given mirror as the effective one.
	Select(mirror Mirror) error
}

// MirrorStore abstracts persistence of the globally selected mirror and any
// custom mirrors added at runtime.
type MirrorStore interface {
	// SelectedMirror returns the currently selected mirror, if any.
	SelectedMirror() (mirror Mirror, ok bool, err error)
	// CustomMirrors returns mirrors that were persisted at runtime.
	CustomMirrors() ([]Mirror, error)
	// Select persists the given mirror as the selected one.
	Select(mirror Mirror) error
}

// ConfigProvider resolves the mirror purely from static configuration. It is
// read-only and used as a fallback when no shared provider is initialized.
type ConfigProvider struct {
	mirror     string
	candidates []Mirror
}

// NewConfigProvider builds a read-only provider from static config.
func NewConfigProvider(mirror string, candidates []Mirror) *ConfigProvider {
	return &ConfigProvider{mirror: mirror, candidates: candidates}
}

// MirrorURL implements PypiProvider.
func (p *ConfigProvider) MirrorURL() string { return p.mirror }

// Candidates implements PypiProvider.
func (p *ConfigProvider) Candidates() []Mirror { return p.candidates }

// CompositeProvider resolves the effective mirror with the following priority:
//  1. the database-selected mirror (if a store is configured and a row exists);
//  2. the statically configured PIP_MIRROR_URL;
//  3. the official PyPI index (represented by an empty MirrorURL()).
//
// When a store is configured it also supports selecting/persisting a mirror.
type CompositeProvider struct {
	configMirror     string
	officialURL      string
	configCandidates []Mirror
	store            MirrorStore // optional; nil disables DB resolution/selection
}

// NewCompositeProvider builds a CompositeProvider. store may be nil.
func NewCompositeProvider(configMirror, officialURL string, candidates []Mirror, store MirrorStore) *CompositeProvider {
	return &CompositeProvider{
		configMirror:     configMirror,
		officialURL:      officialURL,
		configCandidates: candidates,
		store:            store,
	}
}

// MirrorURL implements PypiProvider with DB > config > official priority.
func (p *CompositeProvider) MirrorURL() string {
	if p.store != nil {
		if m, ok, err := p.store.SelectedMirror(); err == nil && ok && m.URL != "" {
			return m.URL
		}
	}
	return p.configMirror
}

// OfficialURL returns the official PyPI index URL used as the default candidate.
func (p *CompositeProvider) OfficialURL() string { return p.officialURL }

// Candidates returns the union of configured candidates, the configured mirror
// and any database-persisted custom mirrors, de-duplicated by URL.
func (p *CompositeProvider) Candidates() []Mirror {
	seen := make(map[string]bool)
	out := make([]Mirror, 0, len(p.configCandidates)+1)

	add := func(m Mirror) {
		m.URL = strings.TrimSpace(m.URL)
		if m.URL == "" || seen[m.URL] {
			return
		}
		seen[m.URL] = true
		out = append(out, m)
	}

	for _, m := range p.configCandidates {
		add(m)
	}
	if p.configMirror != "" {
		add(Mirror{Name: "configured", URL: p.configMirror})
	}
	if p.store != nil {
		if customs, err := p.store.CustomMirrors(); err == nil {
			for _, m := range customs {
				add(m)
			}
		}
	}
	return out
}

// Select persists the chosen mirror. It requires a database-backed store.
func (p *CompositeProvider) Select(mirror Mirror) error {
	if p.store == nil {
		return errors.New("mirror selection requires a database-backed store")
	}
	mirror.URL = strings.TrimSpace(mirror.URL)
	if mirror.URL == "" {
		return errors.New("mirror url is required")
	}
	return p.store.Select(mirror)
}

// ParseCandidates parses a "name=url,name=url" specification into mirrors.
// Entries without a name are accepted (the URL is used as-is).
func ParseCandidates(spec string) []Mirror {
	var out []Mirror
	for _, part := range strings.Split(spec, ",") {
		part = strings.TrimSpace(part)
		if part == "" {
			continue
		}
		name, url, found := strings.Cut(part, "=")
		if !found {
			url = part
			name = ""
		}
		name = strings.TrimSpace(name)
		url = strings.TrimSpace(url)
		if url == "" {
			continue
		}
		out = append(out, Mirror{Name: name, URL: url})
	}
	return out
}

// officialURLOf resolves the official PyPI index URL for a provider, falling
// back to the package default when the provider does not expose one.
func officialURLOf(p PypiProvider) string {
	type officialAware interface{ OfficialURL() string }
	if oa, ok := p.(officialAware); ok {
		if url := oa.OfficialURL(); url != "" {
			return url
		}
	}
	return DefaultPyPIIndexURL
}

// isSelected reports whether a candidate URL corresponds to the effective
// mirror. An empty effective value means the official index is selected.
func isSelected(candidateURL, effective, officialURL string) bool {
	if effective == "" {
		return candidateURL == officialURL
	}
	return candidateURL == effective
}

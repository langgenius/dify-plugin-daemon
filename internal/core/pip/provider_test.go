package pip

import (
	"sync/atomic"
	"testing"

	"github.com/langgenius/dify-plugin-daemon/internal/types/app"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// fakeStore is an in-memory MirrorStore for testing.
type fakeStore struct {
	selected *Mirror
	customs  []Mirror
}

func (s *fakeStore) SelectedMirror() (Mirror, bool, error) {
	if s.selected == nil {
		return Mirror{}, false, nil
	}
	return *s.selected, true, nil
}

func (s *fakeStore) CustomMirrors() ([]Mirror, error) { return s.customs, nil }

func (s *fakeStore) Select(mirror Mirror) error {
	s.selected = &mirror
	for _, m := range s.customs {
		if m.URL == mirror.URL {
			return nil
		}
	}
	s.customs = append(s.customs, mirror)
	return nil
}

func TestParseCandidates(t *testing.T) {
	mirrors := ParseCandidates("aliyun=https://a/simple/, tsinghua=https://t/simple/ , https://x/simple/,")
	require.Len(t, mirrors, 3)
	assert.Equal(t, Mirror{Name: "aliyun", URL: "https://a/simple/"}, mirrors[0])
	assert.Equal(t, Mirror{Name: "tsinghua", URL: "https://t/simple/"}, mirrors[1])
	assert.Equal(t, Mirror{Name: "", URL: "https://x/simple/"}, mirrors[2])
}

func TestCompositeProviderPriorityDBOverConfig(t *testing.T) {
	store := &fakeStore{selected: &Mirror{Name: "db", URL: "https://db/simple/"}}
	provider := NewCompositeProvider("https://config/simple/", "https://pypi/simple/", nil, store)

	// DB selection wins over the configured mirror.
	assert.Equal(t, "https://db/simple/", provider.MirrorURL())
}

func TestCompositeProviderPriorityConfigOverOfficial(t *testing.T) {
	provider := NewCompositeProvider("https://config/simple/", "https://pypi/simple/", nil, &fakeStore{})
	// No DB selection -> the configured mirror is used.
	assert.Equal(t, "https://config/simple/", provider.MirrorURL())
}

func TestCompositeProviderFallsBackToOfficial(t *testing.T) {
	provider := NewCompositeProvider("", "https://pypi/simple/", nil, &fakeStore{})
	// No DB selection and no configured mirror -> empty (official/default index).
	assert.Empty(t, provider.MirrorURL())
}

func TestCompositeProviderCandidatesDedup(t *testing.T) {
	store := &fakeStore{customs: []Mirror{
		{Name: "custom", URL: "https://custom/simple/"},
		{Name: "dup", URL: "https://pypi/simple/"}, // duplicate of official by URL
	}}
	provider := NewCompositeProvider(
		"https://config/simple/",
		"https://pypi/simple/",
		[]Mirror{{Name: "pypi", URL: "https://pypi/simple/"}, {Name: "aliyun", URL: "https://a/simple/"}},
		store,
	)

	candidates := provider.Candidates()
	urls := make([]string, 0, len(candidates))
	for _, c := range candidates {
		urls = append(urls, c.URL)
	}
	assert.Equal(t, []string{
		"https://pypi/simple/",
		"https://a/simple/",
		"https://config/simple/",
		"https://custom/simple/",
	}, urls)
}

func TestCompositeProviderSelectPersists(t *testing.T) {
	store := &fakeStore{}
	provider := NewCompositeProvider("", "https://pypi/simple/", nil, store)

	require.NoError(t, provider.Select(Mirror{Name: "aliyun", URL: "https://a/simple/"}))
	assert.Equal(t, "https://a/simple/", provider.MirrorURL())
}

func TestCompositeProviderSelectRequiresStore(t *testing.T) {
	provider := NewCompositeProvider("", "https://pypi/simple/", nil, nil)
	assert.Error(t, provider.Select(Mirror{URL: "https://a/simple/"}))
}

func TestCompositeProviderSelectRequiresURL(t *testing.T) {
	provider := NewCompositeProvider("", "https://pypi/simple/", nil, &fakeStore{})
	assert.Error(t, provider.Select(Mirror{Name: "nourl"}))
}

func TestBuildMirrorListingMarksSelected(t *testing.T) {
	store := &fakeStore{selected: &Mirror{Name: "aliyun", URL: "https://a/simple/"}}
	provider := NewCompositeProvider(
		"",
		"https://pypi/simple/",
		[]Mirror{{Name: "pypi", URL: "https://pypi/simple/"}, {Name: "aliyun", URL: "https://a/simple/"}},
		store,
	)

	listing := BuildMirrorListing(provider)
	assert.Equal(t, "https://a/simple/", listing.Selected)
	require.Len(t, listing.Mirrors, 2)
	for _, item := range listing.Mirrors {
		if item.URL == "https://a/simple/" {
			assert.True(t, item.Selected)
		} else {
			assert.False(t, item.Selected)
		}
	}
}

func TestProbeEnabled(t *testing.T) {
	assert.True(t, probeEnabled(&app.Config{PipPypiProbeEnabled: true, Platform: app.PLATFORM_LOCAL}))
	assert.False(t, probeEnabled(&app.Config{PipPypiProbeEnabled: false, Platform: app.PLATFORM_LOCAL}))
	assert.False(t, probeEnabled(&app.Config{PipPypiProbeEnabled: true, Platform: app.PLATFORM_SERVERLESS}))
}

func TestProviderOrConfigFallback(t *testing.T) {
	// Ensure no shared provider leaks across tests.
	sharedProvider = atomic.Value{}
	provider := ProviderOrConfig(&app.Config{
		PipMirrorUrl:        "https://config/simple/",
		PipMirrorCandidates: "aliyun=https://a/simple/",
	})
	require.NotNil(t, provider)
	assert.Equal(t, "https://config/simple/", provider.MirrorURL())

	// candidates should include the official index and the configured candidate
	candidates := provider.Candidates()
	require.GreaterOrEqual(t, len(candidates), 2)
}

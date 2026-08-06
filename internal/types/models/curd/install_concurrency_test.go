package curd

import (
	"strings"
	"sync"
	"testing"

	"github.com/google/uuid"
	"github.com/langgenius/dify-plugin-daemon/internal/db"
	"github.com/langgenius/dify-plugin-daemon/internal/types/models"
	"github.com/langgenius/dify-plugin-daemon/pkg/entities/plugin_entities"
	"github.com/stretchr/testify/require"
)

func newConcurrencyTestIdentifier(t *testing.T) plugin_entities.PluginUniqueIdentifier {
	t.Helper()

	pluginName := "concurrency_demo_" + uuid.NewString()
	checksum := uuid.NewString()
	checksum = strings.ReplaceAll(checksum, "-", "")
	if len(checksum) > 32 {
		checksum = checksum[:32]
	}

	identifier, err := plugin_entities.NewPluginUniqueIdentifier("tester/" + pluginName + ":1.0.0.0@" + checksum)
	require.NoError(t, err)
	return identifier
}

func installPluginConcurrently(
	t *testing.T,
	identifier plugin_entities.PluginUniqueIdentifier,
	tenantIDs []string,
) []error {
	t.Helper()

	workers := len(tenantIDs)
	start := make(chan struct{})
	var ready sync.WaitGroup
	var wg sync.WaitGroup
	ready.Add(workers)
	wg.Add(workers)
	errs := make(chan error, workers)
	for _, tenantID := range tenantIDs {
		go func() {
			defer wg.Done()
			ready.Done()
			<-start
			_, _, err := InstallPlugin(
				tenantID,
				identifier,
				plugin_entities.PLUGIN_RUNTIME_TYPE_LOCAL,
				&plugin_entities.PluginDeclaration{},
				"unittest",
				map[string]any{"from": "test"},
			)
			errs <- err
		}()
	}
	ready.Wait()
	close(start)
	wg.Wait()
	close(errs)

	results := make([]error, 0, workers)
	for err := range errs {
		results = append(results, err)
	}
	return results
}

func fetchConcurrencyTestPlugin(
	t *testing.T,
	identifier plugin_entities.PluginUniqueIdentifier,
) models.Plugin {
	t.Helper()

	plugins, err := db.GetAll[models.Plugin](
		db.Equal("plugin_unique_identifier", identifier.String()),
		db.Equal("install_type", string(plugin_entities.PLUGIN_RUNTIME_TYPE_LOCAL)),
	)
	require.NoError(t, err)
	require.Len(t, plugins, 1, "should persist exactly one plugin record")
	return plugins[0]
}

// TestInstallPlugin_IdempotentUnderConcurrency ensures repeated concurrent installation
// for one tenant creates one installation and increments the reference count once.
func TestInstallPlugin_IdempotentUnderConcurrency(t *testing.T) {
	tenantID := uuid.NewString()
	identifier := newConcurrencyTestIdentifier(t)

	const workers = 8
	tenantIDs := make([]string, workers)
	for i := range tenantIDs {
		tenantIDs[i] = tenantID
	}

	errs := installPluginConcurrently(t, identifier, tenantIDs)
	successes := 0
	for _, err := range errs {
		if err == nil {
			successes++
			continue
		}
		require.ErrorIs(t, err, ErrPluginAlreadyInstalled)
	}
	require.Equal(t, 1, successes)

	plugin := fetchConcurrencyTestPlugin(t, identifier)
	require.Equal(t, 1, plugin.Refers)

	installations, err := db.GetAll[models.PluginInstallation](
		db.Equal("plugin_unique_identifier", identifier.String()),
		db.Equal("tenant_id", tenantID),
	)
	require.NoError(t, err)
	require.Len(t, installations, 1, "should persist exactly one installation record for tenant")

	// A subsequent sequential install should be rejected as already installed
	_, _, err = InstallPlugin(
		tenantID,
		identifier,
		plugin_entities.PLUGIN_RUNTIME_TYPE_LOCAL,
		&plugin_entities.PluginDeclaration{},
		"unittest",
		map[string]any{"from": "test"},
	)
	require.ErrorIs(t, err, ErrPluginAlreadyInstalled)
}

// TestInstallPlugin_ConcurrentTenants ensures every tenant can install a plugin while
// the shared plugin record is created and its reference count is updated concurrently.
func TestInstallPlugin_ConcurrentTenants(t *testing.T) {
	identifier := newConcurrencyTestIdentifier(t)

	const workers = 8
	tenantIDs := make([]string, workers)
	for i := range tenantIDs {
		tenantIDs[i] = uuid.NewString()
	}

	for _, err := range installPluginConcurrently(t, identifier, tenantIDs) {
		require.NoError(t, err)
	}

	plugin := fetchConcurrencyTestPlugin(t, identifier)
	require.Equal(t, workers, plugin.Refers)

	installations, err := db.GetAll[models.PluginInstallation](
		db.Equal("plugin_unique_identifier", identifier.String()),
	)
	require.NoError(t, err)
	require.Len(t, installations, workers)

	expectedTenants := make(map[string]struct{}, workers)
	for _, tenantID := range tenantIDs {
		expectedTenants[tenantID] = struct{}{}
	}
	for _, installation := range installations {
		require.Contains(t, expectedTenants, installation.TenantID)
		delete(expectedTenants, installation.TenantID)
	}
	require.Empty(t, expectedTenants)
}

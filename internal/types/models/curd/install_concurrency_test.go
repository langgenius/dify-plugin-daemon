package curd

import (
	"strings"
	"sync"
	"testing"

	"github.com/google/uuid"
	"github.com/langgenius/dify-plugin-daemon/internal/db"
	"github.com/langgenius/dify-plugin-daemon/internal/types/app"
	"github.com/langgenius/dify-plugin-daemon/internal/types/models"
	"github.com/langgenius/dify-plugin-daemon/pkg/entities/plugin_entities"
	"github.com/stretchr/testify/require"
)

// TestInstallPlugin_IdempotentUnderConcurrency ensures creating the same plugin/installation
// concurrently is idempotent: only one plugin and one installation row are persisted.
func TestInstallPlugin_IdempotentUnderConcurrency(t *testing.T) {
	cfg := &app.Config{
		DBType:     app.DB_TYPE_POSTGRESQL,
		DBUsername: "postgres",
		DBPassword: "difyai123456",
		DBHost:     "localhost",
		DBPort:     5432,
		DBDatabase: "dify_plugin_daemon",
		DBSslMode:  "disable",
	}
	cfg.SetDefault()
	db.Init(cfg)
	t.Cleanup(db.Close)

	tenantID := uuid.NewString()
	pluginName := "concurrency_demo_" + uuid.NewString()
	checksum := uuid.NewString()
	checksum = strings.ReplaceAll(checksum, "-", "")
	// 32 hex chars
	if len(checksum) > 32 {
		checksum = checksum[:32]
	}

	identifier, err := plugin_entities.NewPluginUniqueIdentifier("tester/" + pluginName + ":1.0.0.0@" + checksum)
	require.NoError(t, err)

	const workers = 8
	var wg sync.WaitGroup
	wg.Add(workers)
	errs := make(chan error, workers)
	for i := 0; i < workers; i++ {
		go func() {
			defer wg.Done()
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
	wg.Wait()
	close(errs)

	// Validate DB state: exactly one plugin and one installation persisted
	plugins, err := db.GetAll[models.Plugin](
		db.Equal("plugin_unique_identifier", identifier.String()),
		db.Equal("install_type", string(plugin_entities.PLUGIN_RUNTIME_TYPE_LOCAL)),
	)
	require.NoError(t, err)
	require.Len(t, plugins, 1, "should persist exactly one plugin record")

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
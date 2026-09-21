package curd

import (
	"errors"
	"strings"
	"testing"

	"github.com/google/uuid"
	"github.com/langgenius/dify-plugin-daemon/internal/db"
	"github.com/langgenius/dify-plugin-daemon/internal/types/models"
	"github.com/langgenius/dify-plugin-daemon/pkg/entities/plugin_entities"
	"github.com/stretchr/testify/require"
)

func newTestDeclaration(name string) *plugin_entities.PluginDeclaration {
	declaration := &plugin_entities.PluginDeclaration{}
	declaration.Name = name
	return declaration
}

func seedDeclaration(
	t *testing.T,
	identifier plugin_entities.PluginUniqueIdentifier,
	declaration *plugin_entities.PluginDeclaration,
) {
	t.Helper()

	require.NoError(t, db.Create(&models.PluginDeclaration{
		PluginUniqueIdentifier: identifier.String(),
		PluginID:               identifier.PluginID(),
		Declaration:            *declaration,
	}))
}

// declarationName returns the name of the persisted declaration, or "" when there is none.
func declarationName(t *testing.T, identifier plugin_entities.PluginUniqueIdentifier) string {
	t.Helper()

	declaration, err := db.GetOne[models.PluginDeclaration](
		db.Equal("plugin_unique_identifier", identifier.String()),
	)
	if errors.Is(err, db.ErrDatabaseNotFound) {
		return ""
	}
	require.NoError(t, err)
	return declaration.Declaration.Name
}

func TestInstallPlugin_EnsuresDeclaration(t *testing.T) {
	testCases := []struct {
		name         string
		installType  plugin_entities.PluginRuntimeType
		uploadedName string
		wantName     string
	}{
		{
			name:        "creates the missing declaration",
			installType: plugin_entities.PLUGIN_RUNTIME_TYPE_SERVERLESS,
			wantName:    "from-install",
		},
		{
			name:         "keeps the uploaded declaration",
			installType:  plugin_entities.PLUGIN_RUNTIME_TYPE_SERVERLESS,
			uploadedName: "from-upload",
			wantName:     "from-upload",
		},
		{
			name:        "remote plugin gets no declaration row",
			installType: plugin_entities.PLUGIN_RUNTIME_TYPE_REMOTE,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			identifier := newConcurrencyTestIdentifier(t)
			if tc.uploadedName != "" {
				seedDeclaration(t, identifier, newTestDeclaration(tc.uploadedName))
			}

			_, _, err := InstallPlugin(
				uuid.NewString(),
				identifier,
				tc.installType,
				newTestDeclaration("from-install"),
				"unittest",
				map[string]any{"from": "test"},
			)
			require.NoError(t, err)

			require.Equal(t, tc.wantName, declarationName(t, identifier))
		})
	}
}

// The last uninstall deletes the declaration while another install still holds it.
func TestInstallPlugin_RestoresDeletedDeclaration(t *testing.T) {
	identifier := newConcurrencyTestIdentifier(t)
	declaration := newTestDeclaration("from-upload")
	seedDeclaration(t, identifier, declaration)

	firstTenantID := uuid.NewString()
	_, firstInstallation, err := InstallPlugin(
		firstTenantID,
		identifier,
		plugin_entities.PLUGIN_RUNTIME_TYPE_SERVERLESS,
		declaration,
		"unittest",
		map[string]any{"from": "test"},
	)
	require.NoError(t, err)

	response, err := UninstallPlugin(firstTenantID, identifier, firstInstallation.ID, declaration)
	require.NoError(t, err)
	require.True(t, response.IsPluginDeleted)
	require.Empty(t, declarationName(t, identifier))

	_, _, err = InstallPlugin(
		uuid.NewString(),
		identifier,
		plugin_entities.PLUGIN_RUNTIME_TYPE_SERVERLESS,
		declaration,
		"unittest",
		map[string]any{"from": "test"},
	)
	require.NoError(t, err)

	require.Equal(t, "from-upload", declarationName(t, identifier))
}

func TestUpgradePlugin_EnsuresDeclaration(t *testing.T) {
	pluginID := "tester/declaration_demo_" + uuid.NewString()
	identifierFor := func(version string) plugin_entities.PluginUniqueIdentifier {
		checksum := strings.ReplaceAll(uuid.NewString(), "-", "")
		identifier, err := plugin_entities.NewPluginUniqueIdentifier(pluginID + ":" + version + "@" + checksum)
		require.NoError(t, err)
		return identifier
	}
	originalIdentifier, newIdentifier := identifierFor("1.0.0"), identifierFor("1.0.1")
	originalDeclaration := newTestDeclaration("original-version")

	tenantID := uuid.NewString()
	_, _, err := InstallPlugin(
		tenantID,
		originalIdentifier,
		plugin_entities.PLUGIN_RUNTIME_TYPE_SERVERLESS,
		originalDeclaration,
		"unittest",
		map[string]any{"from": "test"},
	)
	require.NoError(t, err)

	_, err = UpgradePlugin(
		tenantID,
		originalIdentifier,
		newIdentifier,
		originalDeclaration,
		newTestDeclaration("new-version"),
		plugin_entities.PLUGIN_RUNTIME_TYPE_SERVERLESS,
		"unittest",
		map[string]any{"from": "test"},
	)
	require.NoError(t, err)

	require.Equal(t, "new-version", declarationName(t, newIdentifier))
}

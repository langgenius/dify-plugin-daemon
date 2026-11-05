package helper

import (
	"testing"

	miniredis "github.com/alicebob/miniredis/v2"
	"github.com/langgenius/dify-plugin-daemon/pkg/entities/plugin_entities"
	"github.com/langgenius/dify-plugin-daemon/pkg/utils/cache"
	"github.com/stretchr/testify/require"
)

func TestCombinedGetPluginDeclarationNormalizesOldCBOR(t *testing.T) {
	server := miniredis.RunT(t)
	require.NoError(t, cache.InitRedisClient(server.Addr(), cache.RedisCredentials{}, false, 0, nil))
	t.Cleanup(func() { _ = cache.Close() })
	id := plugin_entities.PluginUniqueIdentifier("example/defaults:0.0.1@0123456789abcdef")
	runtimeType := plugin_entities.PLUGIN_RUNTIME_TYPE_LOCAL
	t.Cleanup(func() { DeletePluginDeclarationCache(id, runtimeType) })
	for _, models := range [][]plugin_entities.ModelDeclaration{nil, {{Model: "fallback"}}} {
		DeletePluginDeclarationCache(id, runtimeType)
		old := plugin_entities.PluginDeclaration{
			PluginDeclarationWithoutAdvancedFields: plugin_entities.PluginDeclarationWithoutAdvancedFields{
				Description: plugin_entities.I18nObject{EnUS: "description"},
			},
			Model: &plugin_entities.ModelProviderDeclaration{Models: models},
		}
		// Store an unnormalized object exactly as an older CBOR cache would.
		key := "declaration_cache:" + string(runtimeType) + ":" + id.String()
		require.NoError(t, cache.AutoSet(key, old))
		got, err := CombinedGetPluginDeclaration(id, runtimeType)
		require.NoError(t, err)
		require.NotNil(t, got.Tags)
		require.NotNil(t, got.Model.Models)
		require.Equal(t, "description", got.Model.Description.EnUS)
		if len(models) != 0 {
			require.Equal(t, "fallback", got.Model.Models[0].Label.EnUS)
		}
		// Prove the fully prepared object, not the old CBOR value, was published.
		cached, err := CombinedGetPluginDeclaration(id, runtimeType)
		require.NoError(t, err)
		require.Same(t, got, cached)
	}
}

func TestModelInstallationsNormalizesOldCBOR(t *testing.T) {
	server := miniredis.RunT(t)
	require.NoError(t, cache.InitRedisClient(server.Addr(), cache.RedisCredentials{}, false, 0, nil))
	t.Cleanup(func() { _ = cache.Close() })
	wasEnabled := modelInstallationsCacheEnabled.Load()
	SetModelInstallationsCacheEnabled(true)
	t.Cleanup(func() { SetModelInstallationsCacheEnabled(wasEnabled) })
	old := []AIModelInstallationWithDeclaration{
		{Declaration: &plugin_entities.ModelProviderDeclaration{}},
		{Declaration: &plugin_entities.ModelProviderDeclaration{Models: []plugin_entities.ModelDeclaration{{Model: "fallback"}}}},
		{},
	}
	require.NoError(t, cache.SetMapOneField(ModelInstallationsCacheKey("defaults-tenant"), modelInstallationsCacheField(1, 256), old))
	got, err := CombinedListModelInstallations("defaults-tenant", 1, 256)
	require.NoError(t, err)
	require.Len(t, got, 3)
	require.NotNil(t, got[0].Declaration.Models)
	require.Equal(t, "fallback", got[1].Declaration.Models[0].Label.EnUS)
	require.Nil(t, got[2].Declaration)
}

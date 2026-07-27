package helper

import (
	"testing"

	miniredis "github.com/alicebob/miniredis/v2"
	"github.com/langgenius/dify-plugin-daemon/pkg/entities/model_entities"
	"github.com/langgenius/dify-plugin-daemon/pkg/entities/plugin_entities"
	"github.com/langgenius/dify-plugin-daemon/pkg/utils/cache"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const (
	SCHEMA_TEST_IDENTIFIER_V1 = "acme/vertex-connector:0.4.1@0123456789abcdef"
	SCHEMA_TEST_IDENTIFIER_V2 = "acme/vertex-connector:0.5.0@fedcba9876543210"
	SCHEMA_TEST_REMOTE        = "5f8d1c2e-9a3b-4c6d-8e7f-0a1b2c3d4e5f/debug-plugin:0.0.1@aaaabbbbccccdddd"
	SCHEMA_TEST_MODEL_TYPE    = "llm"
	SCHEMA_TEST_MODEL         = "gemini-2.5-pro"
)

func setupModelSchemaTest(t *testing.T) {
	t.Helper()

	redisServer := miniredis.RunT(t)
	require.NoError(t, cache.InitRedisClient(redisServer.Addr(), cache.RedisCredentials{}, false, 0, nil))
	t.Cleanup(func() {
		_ = cache.Close()
	})
}

func schemaTestIdentity(identifier string, credentials map[string]any) ModelSchemaCacheIdentity {
	return ModelSchemaCacheIdentity{
		UniqueIdentifier: plugin_entities.PluginUniqueIdentifier(identifier),
		ModelType:        SCHEMA_TEST_MODEL_TYPE,
		Model:            SCHEMA_TEST_MODEL,
		CredentialType:   "api-key",
		Credentials:      credentials,
	}
}

func schemaWithLabel(label string) model_entities.GetModelSchemasResponse {
	return model_entities.GetModelSchemasResponse{
		ModelSchema: &plugin_entities.ModelDeclaration{
			Model: label,
		},
	}
}

func TestModelSchemaCacheRoundTrip(t *testing.T) {
	setupModelSchemaTest(t)

	identity := schemaTestIdentity(SCHEMA_TEST_IDENTIFIER_V1, map[string]any{"key": "secret-one"})
	assert.Nil(t, GetCachedModelSchema(identity), "nothing is cached before the first store")

	StoreModelSchema(identity, schemaWithLabel("from-cache"))

	cached := GetCachedModelSchema(identity)
	require.NotNil(t, cached)
	require.NotNil(t, cached.ModelSchema)
	assert.Equal(t, "from-cache", cached.ModelSchema.Model)
}

// An upgrade changes the unique identifier, so the old entry is unreachable rather than stale.
func TestModelSchemaCacheIsKeyedOnPluginVersion(t *testing.T) {
	setupModelSchemaTest(t)

	credentials := map[string]any{"key": "secret-one"}
	StoreModelSchema(schemaTestIdentity(SCHEMA_TEST_IDENTIFIER_V1, credentials), schemaWithLabel("old-version"))

	assert.Nil(t,
		GetCachedModelSchema(schemaTestIdentity(SCHEMA_TEST_IDENTIFIER_V2, credentials)),
		"an upgraded plugin must not read the previous version's schema",
	)
}

func TestModelSchemaCacheIsKeyedOnCredentials(t *testing.T) {
	setupModelSchemaTest(t)

	StoreModelSchema(
		schemaTestIdentity(SCHEMA_TEST_IDENTIFIER_V1, map[string]any{"key": "secret-one"}),
		schemaWithLabel("first-credentials"),
	)

	assert.Nil(t,
		GetCachedModelSchema(schemaTestIdentity(SCHEMA_TEST_IDENTIFIER_V1, map[string]any{"key": "secret-two"})),
		"a different credential set must not read the first one's schema",
	)
}

func TestModelSchemaCacheSkipsRemotePlugins(t *testing.T) {
	setupModelSchemaTest(t)

	identity := schemaTestIdentity(SCHEMA_TEST_REMOTE, map[string]any{"key": "secret-one"})
	StoreModelSchema(identity, schemaWithLabel("debug-build"))

	assert.Nil(t, GetCachedModelSchema(identity), "a debugging plugin reuses one identifier across builds")
}

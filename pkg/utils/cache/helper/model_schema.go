package helper

import (
	"crypto/sha256"
	"encoding/hex"
	"strings"
	"sync/atomic"
	"time"

	"github.com/langgenius/dify-plugin-daemon/pkg/entities/model_entities"
	"github.com/langgenius/dify-plugin-daemon/pkg/entities/plugin_entities"
	"github.com/langgenius/dify-plugin-daemon/pkg/utils/cache"
	"github.com/langgenius/dify-plugin-daemon/pkg/utils/log"
	"github.com/langgenius/dify-plugin-daemon/pkg/utils/parser"
)

const DEFAULT_MODEL_SCHEMA_CACHE_TTL = 60 * time.Minute

var modelSchemaCacheTTL atomic.Int64

func SetModelSchemaCacheTTL(ttl time.Duration) {
	if ttl <= 0 {
		ttl = DEFAULT_MODEL_SCHEMA_CACHE_TTL
	}
	modelSchemaCacheTTL.Store(int64(ttl))
}

func currentModelSchemaCacheTTL() time.Duration {
	if ttl := modelSchemaCacheTTL.Load(); ttl > 0 {
		return time.Duration(ttl)
	}
	return DEFAULT_MODEL_SCHEMA_CACHE_TTL
}

type ModelSchemaCacheIdentity struct {
	UniqueIdentifier plugin_entities.PluginUniqueIdentifier
	ModelType        string
	Model            string
	CredentialType   string
	Credentials      map[string]any
}

// Remote and debugging plugins keep one identifier across code changes, so their schemas
// are never cached.
func (i ModelSchemaCacheIdentity) cacheable() bool {
	return !i.UniqueIdentifier.RemoteLike()
}

func (i ModelSchemaCacheIdentity) cacheKey() string {
	digest := sha256.Sum256(parser.MarshalJsonBytes(map[string]any{
		"credential_type": i.CredentialType,
		"credentials":     i.Credentials,
	}))

	return strings.Join(
		[]string{
			"model_schema",
			i.UniqueIdentifier.String(),
			i.ModelType,
			i.Model,
			hex.EncodeToString(digest[:]),
		},
		":",
	)
}

// The key pins the plugin version and checksum, so an upgrade lands on a different key and
// a stale entry is unreachable. Nothing invalidates this cache.
func GetCachedModelSchema(identity ModelSchemaCacheIdentity) *model_entities.GetModelSchemasResponse {
	if !identity.cacheable() {
		return nil
	}

	schema, err := cache.Get[model_entities.GetModelSchemasResponse](identity.cacheKey())
	if err != nil {
		if err != cache.ErrNotFound {
			log.Warn("failed to read model schema cache", "model", identity.Model, "error", err)
		}
		return nil
	}

	return schema
}

func StoreModelSchema(identity ModelSchemaCacheIdentity, schema model_entities.GetModelSchemasResponse) {
	if !identity.cacheable() {
		return
	}

	if err := cache.Store(identity.cacheKey(), schema, currentModelSchemaCacheTTL()); err != nil {
		log.Warn("failed to store model schema cache", "model", identity.Model, "error", err)
	}
}

package plugin_manager

import (
	"fmt"
	"time"

	"github.com/langgenius/dify-plugin-daemon/internal/core/plugin_manager/basic_runtime"
	"github.com/langgenius/dify-plugin-daemon/internal/core/plugin_manager/serverless_runtime"
	"github.com/langgenius/dify-plugin-daemon/internal/db"
	"github.com/langgenius/dify-plugin-daemon/internal/types/models"
	"github.com/langgenius/dify-plugin-daemon/internal/utils/cache"
	"github.com/langgenius/dify-plugin-daemon/internal/utils/cache/helper"
	"github.com/langgenius/dify-plugin-daemon/pkg/entities/plugin_entities"
)

const (
	PLUGIN_SERVERLESS_CACHE_KEY = "serverless:runtime:%s"
)

func (p *PluginManager) getServerlessRuntimeCacheKey(
	identity plugin_entities.PluginUniqueIdentifier,
) string {
	return fmt.Sprintf(PLUGIN_SERVERLESS_CACHE_KEY, identity.String())
}

func (p *PluginManager) getServerlessPluginRuntime(
	identity plugin_entities.PluginUniqueIdentifier,
) (plugin_entities.PluginLifetime, error) {
	model, err := p.getServerlessPluginRuntimeModel(identity)
	if err != nil {
		return nil, err
	}

	// FIXME: get declaration
	declaration, err := helper.CombinedGetPluginDeclaration(identity, plugin_entities.PLUGIN_RUNTIME_TYPE_SERVERLESS)
	if err != nil {
		return nil, err
	}

	// init runtime entity
	runtimeEntity := plugin_entities.PluginRuntime{
		Config: *declaration,
	}
	runtimeEntity.InitState()

	// convert to plugin runtime
	pluginRuntime := serverless_runtime.AWSPluginRuntime{
		BasicChecksum: basic_runtime.BasicChecksum{
			MediaTransport: basic_runtime.NewMediaTransport(p.mediaBucket),
			InnerChecksum:  model.Checksum,
		},
		PluginRuntime: runtimeEntity,
		LambdaURL:     model.FunctionURL,
		LambdaName:    model.FunctionName,
	}

	if err := pluginRuntime.InitEnvironment(); err != nil {
		return nil, err
	}

	return &pluginRuntime, nil
}

func (p *PluginManager) getServerlessPluginRuntimeModel(
	identity plugin_entities.PluginUniqueIdentifier,
) (*models.ServerlessRuntime, error) {
	// check if plugin is a serverless runtime
	runtime, err := cache.Get[models.ServerlessRuntime](
		p.getServerlessRuntimeCacheKey(identity),
	)
	if err != nil && err != cache.ErrNotFound {
		return nil, fmt.Errorf("unexpected error occurred during fetch serverless runtime cache: %v", err)
	}

	if err == cache.ErrNotFound {
		runtimeModel, err := db.GetOne[models.ServerlessRuntime](
			db.Equal("plugin_unique_identifier", identity.String()),
		)

		if err == db.ErrDatabaseNotFound {
			return nil, fmt.Errorf("plugin serverless runtime not found: %s", identity.String())
		}

		if err != nil {
			return nil, fmt.Errorf("failed to load serverless runtime from db: %v", err)
		}

		cache.Store(p.getServerlessRuntimeCacheKey(identity), runtimeModel, time.Minute*30)
		runtime = &runtimeModel
	} else if err != nil {
		return nil, fmt.Errorf("failed to load serverless runtime from cache: %v", err)
	}

	return runtime, nil
}

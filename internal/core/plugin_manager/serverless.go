package plugin_manager

import (
	"errors"
	"fmt"
	"time"

	"github.com/langgenius/dify-plugin-daemon/internal/core/plugin_manager/aws_manager"
	"github.com/langgenius/dify-plugin-daemon/internal/core/plugin_manager/basic_manager"
	"github.com/langgenius/dify-plugin-daemon/internal/core/plugin_manager/positive_manager"
	"github.com/langgenius/dify-plugin-daemon/internal/db"
	"github.com/langgenius/dify-plugin-daemon/internal/types/entities/plugin_entities"
	"github.com/langgenius/dify-plugin-daemon/internal/types/models"
	"github.com/langgenius/dify-plugin-daemon/internal/utils/cache"
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

	declaration, err := model.GetDeclaration()
	if err != nil {
		return nil, err
	}

	// init runtime entity
	runtime_entity := plugin_entities.PluginRuntime{
		Config: *declaration,
	}
	runtime_entity.InitState()

	// convert to plugin runtime
	plugin_runtime := aws_manager.AWSPluginRuntime{
		PositivePluginRuntime: positive_manager.PositivePluginRuntime{
			BasicPluginRuntime: basic_manager.NewBasicPluginRuntime(p.mediaManager),
			InnerChecksum:      model.Checksum,
		},
		PluginRuntime: runtime_entity,
		LambdaURL:     model.FunctionURL,
		LambdaName:    model.FunctionName,
	}

	if err := plugin_runtime.InitEnvironment(); err != nil {
		return nil, err
	}

	return &plugin_runtime, nil
}

func (p *PluginManager) getServerlessPluginRuntimeModel(
	identity plugin_entities.PluginUniqueIdentifier,
) (*models.ServerlessRuntime, error) {
	// check if plugin is a serverless runtime
	runtime, err := cache.Get[models.ServerlessRuntime](
		p.getServerlessRuntimeCacheKey(identity),
	)
	if err != nil && err != cache.ErrNotFound {
		return nil, errors.New("plugin not found")
	}

	if err == cache.ErrNotFound {
		runtime_model, err := db.GetOne[models.ServerlessRuntime](
			db.Equal("plugin_unique_identifier", identity.String()),
		)

		if err == db.ErrDatabaseNotFound {
			return nil, errors.New("plugin not found")
		}

		if err != nil {
			return nil, fmt.Errorf("failed to load serverless runtime from db: %v", err)
		}

		cache.Store(p.getServerlessRuntimeCacheKey(identity), runtime_model, time.Minute*30)
		runtime = &runtime_model
	} else if err != nil {
		return nil, fmt.Errorf("failed to load serverless runtime from cache: %v", err)
	}

	return runtime, nil
}

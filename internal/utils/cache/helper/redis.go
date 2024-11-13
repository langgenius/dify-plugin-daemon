package helper

import (
	"strings"

	"github.com/langgenius/dify-plugin-daemon/internal/db"
	"github.com/langgenius/dify-plugin-daemon/internal/types/entities/plugin_entities"
	"github.com/langgenius/dify-plugin-daemon/internal/types/models"
	"github.com/langgenius/dify-plugin-daemon/internal/utils/cache"
)

func CombinedGetPluginDeclaration(
	plugin_unique_identifier plugin_entities.PluginUniqueIdentifier,
	tenant_id string,
	runtime_type plugin_entities.PluginRuntimeType,
) (*plugin_entities.PluginDeclaration, error) {
	return cache.AutoGetWithGetter(
		strings.Join(
			[]string{
				string(runtime_type),
				tenant_id,
				plugin_unique_identifier.String(),
			},
			":",
		),
		func() (*plugin_entities.PluginDeclaration, error) {
			if runtime_type != plugin_entities.PLUGIN_RUNTIME_TYPE_REMOTE {
				declaration, err := db.GetOne[models.PluginDeclaration](
					db.Equal("plugin_unique_identifier", plugin_unique_identifier.String()),
				)
				if err != nil && err != db.ErrDatabaseNotFound {
					return nil, err
				}

				if err != nil {
					return nil, err
				}

				return &declaration.Declaration, nil
			} else {
				// try to fetch the declaration from plugin if it's remote
				plugin, err := db.GetOne[models.Plugin](
					db.Equal("unique_identifier", plugin_unique_identifier.String()),
					db.Equal("install_type", string(plugin_entities.PLUGIN_RUNTIME_TYPE_REMOTE)),
					db.Equal("tenant_id", tenant_id),
				)
				if err != nil && err != db.ErrDatabaseNotFound {
					return nil, err
				}

				if err != nil {
					return nil, err
				}

				return &plugin.Declaration, nil
			}
		},
	)
}

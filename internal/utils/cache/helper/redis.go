package helper

import (
	"errors"
	"strings"

	"github.com/langgenius/dify-plugin-daemon/internal/db"
	"github.com/langgenius/dify-plugin-daemon/internal/types/models"
	"github.com/langgenius/dify-plugin-daemon/internal/utils/cache"
	"github.com/langgenius/dify-plugin-daemon/pkg/entities/plugin_entities"
)

var (
	ErrPluginNotFound = errors.New("plugin not found")
)

func CombinedGetPluginDeclaration(
	pluginUniqueIdentifier plugin_entities.PluginUniqueIdentifier,
	tenantId string,
	runtimeType plugin_entities.PluginRuntimeType,
) (*plugin_entities.PluginDeclaration, error) {
	return cache.AutoGetWithGetter(
		strings.Join(
			[]string{
				string(runtimeType),
				pluginUniqueIdentifier.String(),
			},
			":",
		),
		func() (*plugin_entities.PluginDeclaration, error) {
			if runtimeType != plugin_entities.PLUGIN_RUNTIME_TYPE_REMOTE {
				declaration, err := db.GetOne[models.PluginDeclaration](
					db.Equal("plugin_unique_identifier", pluginUniqueIdentifier.String()),
				)
				if err == db.ErrDatabaseNotFound {
					return nil, ErrPluginNotFound
				}

				if err != nil {
					return nil, err
				}

				return &declaration.Declaration, nil
			} else {
				// try to fetch the declaration from plugin if it's remote
				plugin, err := db.GetOne[models.Plugin](
					db.Equal("plugin_unique_identifier", pluginUniqueIdentifier.String()),
					db.Equal("install_type", string(plugin_entities.PLUGIN_RUNTIME_TYPE_REMOTE)),
				)
				if err == db.ErrDatabaseNotFound {
					return nil, ErrPluginNotFound
				}

				if err != nil {
					return nil, err
				}

				return &plugin.Declaration, nil
			}
		},
	)
}

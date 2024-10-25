package helper

import (
	"github.com/langgenius/dify-plugin-daemon/internal/db"
	"github.com/langgenius/dify-plugin-daemon/internal/types/entities/plugin_entities"
	"github.com/langgenius/dify-plugin-daemon/internal/types/models"
	"github.com/langgenius/dify-plugin-daemon/internal/utils/cache"
)

func CombinedGetPluginDeclaration(plugin_unique_identifier plugin_entities.PluginUniqueIdentifier) (*plugin_entities.PluginDeclaration, error) {
	return cache.AutoGetWithGetter(
		plugin_unique_identifier.String(),
		func() (*plugin_entities.PluginDeclaration, error) {
			declaration, err := db.GetOne[models.PluginDeclaration](
				db.Equal("plugin_unique_identifier", plugin_unique_identifier.String()),
			)
			if err != nil && err != db.ErrDatabaseNotFound {
				return nil, err
			}

			if err == nil {
				return &declaration.Declaration, nil
			}

			model, err := db.GetOne[models.Plugin](
				db.Equal("plugin_unique_identifier", plugin_unique_identifier.String()),
			)
			if err != nil {
				return nil, err
			}

			return &model.Declaration, nil
		},
	)
}

package service

import (
	"github.com/langgenius/dify-plugin-daemon/internal/db"
	"github.com/langgenius/dify-plugin-daemon/internal/types/exception"
	"github.com/langgenius/dify-plugin-daemon/internal/types/models"
	"github.com/langgenius/dify-plugin-daemon/internal/utils/cache/helper"
	"github.com/langgenius/dify-plugin-daemon/pkg/entities"
	"github.com/langgenius/dify-plugin-daemon/pkg/entities/plugin_entities"
)

func ListTriggers(tenant_id string, page int, page_size int) *entities.Response {
	type Trigger struct {
		models.TriggerInstallation
		Declaration *plugin_entities.TriggerProviderDeclaration `json:"declaration"`
	}

	triggers, err := db.GetAll[models.TriggerInstallation](
		db.Equal("tenant_id", tenant_id),
		db.Page(page, page_size),
	)

	if err != nil {
		return exception.InternalServerError(err).ToResponse()
	}

	data := make([]Trigger, 0, len(triggers))

	for _, trigger := range triggers {
		uniqueIdentifier := plugin_entities.PluginUniqueIdentifier(trigger.PluginUniqueIdentifier)
		var runtimeType plugin_entities.PluginRuntimeType
		if uniqueIdentifier.RemoteLike() {
			runtimeType = plugin_entities.PLUGIN_RUNTIME_TYPE_REMOTE
		} else {
			runtimeType = plugin_entities.PLUGIN_RUNTIME_TYPE_LOCAL
		}

		declaration, err := helper.CombinedGetPluginDeclaration(
			uniqueIdentifier,
			runtimeType,
		)

		if err != nil {
			return exception.InternalServerError(err).ToResponse()
		}

		data = append(data, Trigger{
			TriggerInstallation: trigger,
			Declaration:         declaration.Trigger,
		})
	}

	return entities.NewSuccessResponse(data)
}

func GetTrigger(tenant_id string, plugin_id string, provider string) *entities.Response {
	type Trigger struct {
		models.TriggerInstallation
		Declaration *plugin_entities.TriggerProviderDeclaration `json:"declaration"`
	}

	trigger, err := db.GetOne[models.TriggerInstallation](
		db.Equal("tenant_id", tenant_id),
		db.Equal("plugin_id", plugin_id),
	)

	if err != nil {
		if err == db.ErrDatabaseNotFound {
			return exception.ErrPluginNotFound().ToResponse()
		}

		return exception.InternalServerError(err).ToResponse()
	}

	if trigger.Provider != provider {
		return exception.ErrPluginNotFound().ToResponse()
	}

	uniqueIdentifier := plugin_entities.PluginUniqueIdentifier(trigger.PluginUniqueIdentifier)
	var runtimeType plugin_entities.PluginRuntimeType
	if uniqueIdentifier.RemoteLike() {
		runtimeType = plugin_entities.PLUGIN_RUNTIME_TYPE_REMOTE
	} else {
		runtimeType = plugin_entities.PLUGIN_RUNTIME_TYPE_LOCAL
	}

	declaration, err := helper.CombinedGetPluginDeclaration(
		uniqueIdentifier,
		runtimeType,
	)

	if err != nil {
		return exception.InternalServerError(err).ToResponse()
	}

	return entities.NewSuccessResponse(Trigger{
		TriggerInstallation: trigger,
		Declaration:         declaration.Trigger,
	})
}
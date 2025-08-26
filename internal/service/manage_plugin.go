package service

import (
	"errors"
	"time"

	"github.com/langgenius/dify-plugin-daemon/internal/db"
	"github.com/langgenius/dify-plugin-daemon/internal/types/exception"
	"github.com/langgenius/dify-plugin-daemon/internal/types/models"
	"github.com/langgenius/dify-plugin-daemon/internal/utils/cache/helper"
	"github.com/langgenius/dify-plugin-daemon/internal/utils/strings"
	"github.com/langgenius/dify-plugin-daemon/pkg/entities"
	"github.com/langgenius/dify-plugin-daemon/pkg/entities/manifest_entities"
	"github.com/langgenius/dify-plugin-daemon/pkg/entities/plugin_entities"
)

func ListPlugins(tenant_id string, page int, page_size int) *entities.Response {
	type installation struct {
		ID                     string                             `json:"id"`
		Name                   string                             `json:"name"`
		PluginID               string                             `json:"plugin_id"`
		TenantID               string                             `json:"tenant_id"`
		PluginUniqueIdentifier string                             `json:"plugin_unique_identifier"`
		EndpointsActive        int                                `json:"endpoints_active"`
		EndpointsSetups        int                                `json:"endpoints_setups"`
		InstallationID         string                             `json:"installation_id"`
		Declaration            *plugin_entities.PluginDeclaration `json:"declaration"`
		RuntimeType            plugin_entities.PluginRuntimeType  `json:"runtime_type"`
		Version                manifest_entities.Version          `json:"version"`
		CreatedAt              time.Time                          `json:"created_at"`
		UpdatedAt              time.Time                          `json:"updated_at"`
		Source                 string                             `json:"source"`
		Checksum               string                             `json:"checksum"`
		Meta                   map[string]any                     `json:"meta"`
	}

	type responseData struct {
		List  []installation `json:"list"`
		Total int64          `json:"total"`
	}

	// get total count
	totalCount, err := db.GetCount[models.PluginInstallation](
		db.Equal("tenant_id", tenant_id),
	)

	if err != nil {
		return exception.InternalServerError(err).ToResponse()
	}

	pluginInstallations, err := db.GetAll[models.PluginInstallation](
		db.Equal("tenant_id", tenant_id),
		db.OrderBy("created_at", true),
		db.Page(page, page_size),
	)

	if err != nil {
		return exception.InternalServerError(err).ToResponse()
	}

	data := make([]installation, 0, len(pluginInstallations))

	for _, plugin_installation := range pluginInstallations {
		pluginUniqueIdentifier, err := plugin_entities.NewPluginUniqueIdentifier(
			plugin_installation.PluginUniqueIdentifier,
		)
		if err != nil {
			return exception.UniqueIdentifierError(err).ToResponse()
		}

		pluginDeclaration, err := helper.CombinedGetPluginDeclaration(
			pluginUniqueIdentifier,
			plugin_entities.PluginRuntimeType(plugin_installation.RuntimeType),
		)
		if err != nil {
			return exception.InternalServerError(err).ToResponse()
		}

		data = append(data, installation{
			ID:                     plugin_installation.ID,
			Name:                   pluginDeclaration.Name,
			TenantID:               plugin_installation.TenantID,
			PluginID:               pluginUniqueIdentifier.PluginID(),
			PluginUniqueIdentifier: pluginUniqueIdentifier.String(),
			InstallationID:         plugin_installation.ID,
			Declaration:            pluginDeclaration,
			EndpointsSetups:        plugin_installation.EndpointsSetups,
			EndpointsActive:        plugin_installation.EndpointsActive,
			RuntimeType:            plugin_entities.PluginRuntimeType(plugin_installation.RuntimeType),
			Version:                pluginDeclaration.Version,
			CreatedAt:              plugin_installation.CreatedAt,
			UpdatedAt:              plugin_installation.UpdatedAt,
			Source:                 plugin_installation.Source,
			Meta:                   plugin_installation.Meta,
			Checksum:               pluginUniqueIdentifier.Checksum(),
		})
	}

	finalData := responseData{
		List: 	data,
		Total: 	totalCount,
	}

	return entities.NewSuccessResponse(finalData)
}

// Using plugin_ids to fetch plugin installations
func BatchFetchPluginInstallationByIDs(tenant_id string, plugin_ids []string) *entities.Response {
	type installation struct {
		models.PluginInstallation

		Version     manifest_entities.Version          `json:"version"`
		Checksum    string                             `json:"checksum"`
		Declaration *plugin_entities.PluginDeclaration `json:"declaration"`
	}

	if len(plugin_ids) == 0 {
		return entities.NewSuccessResponse([]installation{})
	}

	pluginInstallations, err := db.GetAll[models.PluginInstallation](
		db.Equal("tenant_id", tenant_id),
		db.InArray("plugin_id", strings.Map(plugin_ids, func(id string) any { return id })),
		db.Page(1, 256), // TODO: pagination
	)

	if err != nil {
		return exception.InternalServerError(err).ToResponse()
	}

	data := make([]installation, 0, len(pluginInstallations))

	for _, plugin_installation := range pluginInstallations {
		pluginUniqueIdentifier, err := plugin_entities.NewPluginUniqueIdentifier(
			plugin_installation.PluginUniqueIdentifier,
		)

		if err != nil {
			return exception.InternalServerError(errors.Join(errors.New("invalid plugin unique identifier found"), err)).ToResponse()
		}

		pluginDeclaration, err := helper.CombinedGetPluginDeclaration(
			pluginUniqueIdentifier,
			plugin_entities.PluginRuntimeType(plugin_installation.RuntimeType),
		)
		if err != nil {
			return exception.InternalServerError(errors.Join(errors.New("failed to get plugin declaration"), err)).ToResponse()
		}

		data = append(data, installation{
			PluginInstallation: plugin_installation,
			Version:            pluginUniqueIdentifier.Version(),
			Checksum:           pluginUniqueIdentifier.Checksum(),
			Declaration:        pluginDeclaration,
		})
	}

	return entities.NewSuccessResponse(data)
}

// check which plugin is missing
func FetchMissingPluginInstallations(tenant_id string, plugin_unique_identifiers []plugin_entities.PluginUniqueIdentifier) *entities.Response {
	type MissingPluginDependency struct {
		PluginUniqueIdentifier string `json:"plugin_unique_identifier"`
		CurrentIdentifier      string `json:"current_identifier"`
	}
	result := make([]MissingPluginDependency, 0, len(plugin_unique_identifiers))

	if len(plugin_unique_identifiers) == 0 {
		return entities.NewSuccessResponse(result)
	}

	installed, err := db.GetAll[models.PluginInstallation](
		db.Equal("tenant_id", tenant_id),
		db.InArray(
			"plugin_id",
			strings.Map(
				plugin_unique_identifiers,
				func(id plugin_entities.PluginUniqueIdentifier) any {
					return id.PluginID()
				},
			),
		),
		db.Page(1, 256), // TODO: pagination
	)

	if err != nil {
		return exception.InternalServerError(err).ToResponse()
	}

	// check which plugin is missing
	for _, pluginUniqueIdentifier := range plugin_unique_identifiers {
		found := false
		for _, installedPlugin := range installed {
			if installedPlugin.PluginID == pluginUniqueIdentifier.PluginID() {
				found = true
				if installedPlugin.PluginUniqueIdentifier != pluginUniqueIdentifier.String() {
					// version mismatched
					result = append(result, MissingPluginDependency{
						PluginUniqueIdentifier: pluginUniqueIdentifier.String(),
						CurrentIdentifier:      installedPlugin.PluginUniqueIdentifier,
					})
				}
				break
			}
		}

		if !found {
			result = append(result, MissingPluginDependency{
				PluginUniqueIdentifier: pluginUniqueIdentifier.String(),
			})
		}
	}

	return entities.NewSuccessResponse(result)
}

func ListTools(tenant_id string, page int, page_size int) *entities.Response {
	type Tool struct {
		models.ToolInstallation // pointer to avoid deep copy

		Declaration *plugin_entities.ToolProviderDeclaration `json:"declaration"`
	}

	providers, err := db.GetAll[models.ToolInstallation](
		db.Equal("tenant_id", tenant_id),
		db.Page(page, page_size),
	)

	if err != nil {
		return exception.InternalServerError(err).ToResponse()
	}

	data := make([]Tool, 0, len(providers))

	for _, provider := range providers {
		// check if plugin id starts with uuid
		// split by uuid length
		uniqueIdentifier := plugin_entities.PluginUniqueIdentifier(provider.PluginUniqueIdentifier)
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

		data = append(data, Tool{
			ToolInstallation: provider,
			Declaration:      declaration.Tool,
		})
	}

	return entities.NewSuccessResponse(data)
}

func ListModels(tenant_id string, page int, page_size int) *entities.Response {
	type AIModel struct {
		models.AIModelInstallation // pointer to avoid deep copy

		Declaration *plugin_entities.ModelProviderDeclaration `json:"declaration"`
	}

	providers, err := db.GetAll[models.AIModelInstallation](
		db.Equal("tenant_id", tenant_id),
		db.Page(page, page_size),
	)

	if err != nil {
		return exception.InternalServerError(err).ToResponse()
	}

	data := make([]AIModel, 0, len(providers))

	for _, provider := range providers {
		uniqueIdentifier := plugin_entities.PluginUniqueIdentifier(provider.PluginUniqueIdentifier)
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

		data = append(data, AIModel{
			AIModelInstallation: provider,
			Declaration:         declaration.Model,
		})
	}

	return entities.NewSuccessResponse(data)
}

func GetTool(tenant_id string, plugin_id string, provider string) *entities.Response {
	type Tool struct {
		models.ToolInstallation // pointer to avoid deep copy

		Declaration *plugin_entities.ToolProviderDeclaration `json:"declaration"`
	}

	// try get tool
	tool, err := db.GetOne[models.ToolInstallation](
		db.Equal("tenant_id", tenant_id),
		db.Equal("plugin_id", plugin_id),
	)

	if err != nil {
		if err == db.ErrDatabaseNotFound {
			return exception.ErrPluginNotFound().ToResponse()
		}

		return exception.InternalServerError(err).ToResponse()
	}

	if tool.Provider != provider {
		return exception.ErrPluginNotFound().ToResponse()
	}

	uniqueIdentifier := plugin_entities.PluginUniqueIdentifier(tool.PluginUniqueIdentifier)
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

	return entities.NewSuccessResponse(Tool{
		ToolInstallation: tool,
		Declaration:      declaration.Tool,
	})
}

type RequestCheckToolExistence struct {
	PluginID     string `json:"plugin_id" validate:"required"`
	ProviderName string `json:"provider_name" validate:"required"`
}

func CheckToolExistence(tenantId string, providerIds []RequestCheckToolExistence) *entities.Response {
	existence := make([]bool, 0, len(providerIds))

	// get all providers
	providers, err := db.GetAll[models.ToolInstallation](
		db.Equal("tenant_id", tenantId),
		db.InArray("plugin_id", strings.Map(providerIds, func(id RequestCheckToolExistence) any { return id.PluginID })),
		db.Page(1, 256), // TODO: pagination
	)

	if err != nil {
		return exception.InternalServerError(err).ToResponse()
	}

	// check provider id
	for _, providerId := range providerIds {
		found := false
		for _, provider := range providers {
			if provider.PluginID == providerId.PluginID && provider.Provider == providerId.ProviderName {
				found = true
				break
			}
		}

		existence = append(existence, found)
	}

	return entities.NewSuccessResponse(existence)
}

func ListAgentStrategies(tenant_id string, page int, page_size int) *entities.Response {
	type AgentStrategy struct {
		models.AgentStrategyInstallation // pointer to avoid deep copy

		Declaration *plugin_entities.AgentStrategyProviderDeclaration `json:"declaration"`
		Meta        plugin_entities.PluginMeta                        `json:"meta"`
	}

	providers, err := db.GetAll[models.AgentStrategyInstallation](
		db.Equal("tenant_id", tenant_id),
		db.Page(page, page_size),
	)

	if err != nil {
		return exception.InternalServerError(err).ToResponse()
	}

	data := make([]AgentStrategy, 0, len(providers))

	for _, provider := range providers {
		uniqueIdentifier := plugin_entities.PluginUniqueIdentifier(provider.PluginUniqueIdentifier)
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

		data = append(data, AgentStrategy{
			AgentStrategyInstallation: provider,
			Declaration:               declaration.AgentStrategy,
			Meta:                      declaration.Meta,
		})
	}

	return entities.NewSuccessResponse(data)
}

func GetAgentStrategy(tenant_id string, plugin_id string, provider string) *entities.Response {
	type AgentStrategy struct {
		models.AgentStrategyInstallation // pointer to avoid deep copy

		Declaration *plugin_entities.AgentStrategyProviderDeclaration `json:"declaration"`
		Meta        plugin_entities.PluginMeta                        `json:"meta"`
	}

	agent_strategy, err := db.GetOne[models.AgentStrategyInstallation](
		db.Equal("tenant_id", tenant_id),
		db.Equal("plugin_id", plugin_id),
	)

	if err != nil {
		if err == db.ErrDatabaseNotFound {
			return exception.ErrPluginNotFound().ToResponse()
		}

		return exception.InternalServerError(err).ToResponse()
	}

	if agent_strategy.Provider != provider {
		return exception.ErrPluginNotFound().ToResponse()
	}

	uniqueIdentifier := plugin_entities.PluginUniqueIdentifier(agent_strategy.PluginUniqueIdentifier)
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

	return entities.NewSuccessResponse(AgentStrategy{
		AgentStrategyInstallation: agent_strategy,
		Declaration:               declaration.AgentStrategy,
		Meta:                      declaration.Meta,
	})
}

func ListDatasources(tenant_id string, page int, page_size int) *entities.Response {
	type Datasource struct {
		models.DatasourceInstallation // pointer to avoid deep copy

		Declaration *plugin_entities.DatasourceProviderDeclaration `json:"declaration"`
	}

	providers, err := db.GetAll[models.DatasourceInstallation](
		db.Equal("tenant_id", tenant_id),
		db.Page(page, page_size),
	)

	if err != nil {
		return exception.InternalServerError(err).ToResponse()
	}

	data := make([]Datasource, 0, len(providers))

	for _, provider := range providers {
		uniqueIdentifier := plugin_entities.PluginUniqueIdentifier(provider.PluginUniqueIdentifier)
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

		data = append(data, Datasource{
			DatasourceInstallation: provider,
			Declaration:            declaration.Datasource,
		})
	}

	return entities.NewSuccessResponse(data)
}

func GetDatasource(tenant_id string, plugin_id string, provider string) *entities.Response {
	type Datasource struct {
		models.DatasourceInstallation // pointer to avoid deep copy

		Declaration *plugin_entities.DatasourceProviderDeclaration `json:"declaration"`
	}

	datasource, err := db.GetOne[models.DatasourceInstallation](
		db.Equal("tenant_id", tenant_id),
		db.Equal("plugin_id", plugin_id),
	)

	if err != nil {
		return exception.InternalServerError(err).ToResponse()
	}

	if datasource.Provider != provider {
		return exception.ErrPluginNotFound().ToResponse()
	}

	uniqueIdentifier := plugin_entities.PluginUniqueIdentifier(datasource.PluginUniqueIdentifier)
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

	return entities.NewSuccessResponse(Datasource{
		DatasourceInstallation: datasource,
		Declaration:            declaration.Datasource,
	})
}

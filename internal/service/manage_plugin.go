package service

import (
	"time"

	"github.com/langgenius/dify-plugin-daemon/internal/db"
	"github.com/langgenius/dify-plugin-daemon/internal/types/entities"
	"github.com/langgenius/dify-plugin-daemon/internal/types/entities/plugin_entities"
	"github.com/langgenius/dify-plugin-daemon/internal/types/models"
	"github.com/langgenius/dify-plugin-daemon/internal/utils/cache"
)

func ListPlugins(tenant_id string, page int, page_size int) *entities.Response {
	type installation struct {
		ID             string                             `json:"id"`
		Name           string                             `json:"name"`
		PluginID       string                             `json:"plugin_id"`
		InstallationID string                             `json:"installation_id"`
		Description    *plugin_entities.PluginDeclaration `json:"description"`
		RuntimeType    plugin_entities.PluginRuntimeType  `json:"runtime_type"`
		Version        string                             `json:"version"`
		CreatedAt      time.Time                          `json:"created_at"`
		UpdatedAt      time.Time                          `json:"updated_at"`
	}

	plugin_installations, err := db.GetAll[models.PluginInstallation](
		db.Equal("tenant_id", tenant_id),
		db.Page(page, page_size),
	)

	if err != nil {
		return entities.NewErrorResponse(-500, err.Error())
	}

	data := make([]installation, 0, len(plugin_installations))

	for _, plugin_installation := range plugin_installations {
		plugin, err := cache.Get[models.Plugin](plugin_installation.PluginUniqueIdentifier)
		if err != nil {
			return entities.NewErrorResponse(-500, err.Error())
		}

		declaration := plugin.Declaration
		data = append(data, installation{
			ID:             plugin_installation.ID,
			Name:           declaration.Name,
			PluginID:       plugin.ID,
			InstallationID: plugin_installation.ID,
			Description:    &declaration,
			RuntimeType:    plugin_entities.PluginRuntimeType(plugin_installation.RuntimeType),
			Version:        declaration.Version,
			CreatedAt:      plugin_installation.CreatedAt,
			UpdatedAt:      plugin_installation.UpdatedAt,
		})
	}

	return entities.NewSuccessResponse(data)
}

func ListTools(tenant_id string, page int, page_size int) *entities.Response {
	return nil
}

func ListModels(tenant_id string, page int, page_size int) *entities.Response {
	providers := make([]plugin_entities.ModelProviderDeclaration, 0)
	providers = append(providers, plugin_entities.ModelProviderDeclaration{
		Provider: "openai",
		Label: plugin_entities.I18nObject{
			EnUS:   "OpenAI",
			ZhHans: "OpenAI",
			JaJp:   "OpenAI",
			PtBr:   "OpenAI",
		},
		Description: &plugin_entities.I18nObject{
			EnUS:   "OpenAI",
			ZhHans: "OpenAI",
			JaJp:   "OpenAI",
			PtBr:   "OpenAI",
		},
		IconSmall: &plugin_entities.I18nObject{
			EnUS:   "icon_small.svg",
			ZhHans: "icon_small.svg",
			JaJp:   "icon_small.svg",
			PtBr:   "icon_small.svg",
		},
		IconLarge: &plugin_entities.I18nObject{
			EnUS:   "icon_large.svg",
			ZhHans: "icon_large.svg",
			JaJp:   "icon_large.svg",
			PtBr:   "icon_large.svg",
		},
		Background: &[]string{"background.svg"}[0],
		SupportedModelTypes: []plugin_entities.ModelType{
			plugin_entities.MODEL_TYPE_LLM,
		},
		ConfigurateMethods: []plugin_entities.ModelProviderConfigurateMethod{
			plugin_entities.CONFIGURATE_METHOD_PREDEFINED_MODEL,
			plugin_entities.CONFIGURATE_METHOD_CUSTOMIZABLE_MODEL,
		},
		ProviderCredentialSchema: &plugin_entities.ModelProviderCredentialSchema{
			CredentialFormSchemas: []plugin_entities.ModelProviderCredentialFormSchema{
				{
					Variable: "api_key",
					Label: plugin_entities.I18nObject{
						EnUS:   "API Key",
						ZhHans: "API Key",
						JaJp:   "API Key",
						PtBr:   "API Key",
					},
					Type:      plugin_entities.FORM_TYPE_SECRET_INPUT,
					Required:  true,
					MaxLength: 1024,
				},
			},
		},
		ModelDeclarations: []plugin_entities.ModelDeclaration{
			{
				Model: "gpt-4o",
				Label: plugin_entities.I18nObject{
					EnUS:   "GPT-4o",
					ZhHans: "GPT-4o",
					JaJp:   "GPT-4o",
					PtBr:   "GPT-4o",
				},
				ModelType: plugin_entities.MODEL_TYPE_LLM,
				Features: []string{
					"multi-tool-call",
				},
				FetchFrom: plugin_entities.CONFIGURATE_METHOD_PREDEFINED_MODEL,
				ModelProperties: map[string]any{
					"mode":         "chat",
					"context_size": 128000,
				},
				ParameterRules: []plugin_entities.ModelParameterRule{
					{
						Name: "temperature",
						Label: &plugin_entities.I18nObject{
							EnUS:   "Temperature",
							ZhHans: "温度",
							JaJp:   "温度",
							PtBr:   "温度",
						},
						Type:      &[]plugin_entities.ModelParameterType{plugin_entities.PARAMETER_TYPE_FLOAT}[0],
						Required:  true,
						Min:       &[]float64{0}[0],
						Max:       &[]float64{1}[0],
						Default:   &[]any{0.7}[0],
						Precision: &[]int{2}[0],
					},
				},
			},
		},
	})

	return entities.NewSuccessResponse(providers)
}

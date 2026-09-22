package controllers

import (
	"github.com/langgenius/dify-plugin-daemon/internal/server/contracts"
	"github.com/langgenius/dify-plugin-daemon/internal/service"
	"github.com/langgenius/dify-plugin-daemon/pkg/entities/plugin_entities"
)

func toolProviderInstallation(tool service.InstalledTool) contracts.ToolProviderInstallation {
	return contracts.ToolProviderInstallation{
		ID:                     tool.ID,
		CreatedAt:              tool.CreatedAt,
		UpdatedAt:              tool.UpdatedAt,
		TenantID:               tool.TenantID,
		Provider:               tool.Provider,
		PluginUniqueIdentifier: tool.PluginUniqueIdentifier,
		PluginID:               tool.PluginID,
		Declaration:            toolProviderDeclaration(tool.Declaration),
	}
}

func toolProviderDeclaration(declaration *plugin_entities.ToolProviderDeclaration) *contracts.ToolProviderDeclaration {
	if declaration == nil {
		return nil
	}
	var tags []contracts.PluginTag
	if declaration.Identity.Tags != nil {
		tags = make([]contracts.PluginTag, len(declaration.Identity.Tags))
		for i, tag := range declaration.Identity.Tags {
			tags[i] = contracts.PluginTag(tag)
		}
	}
	credentials := toolProviderConfigs(declaration.CredentialsSchema)
	if credentials == nil {
		credentials = []contracts.ProviderConfig{}
	}
	tools := make([]contracts.ToolDeclaration, len(declaration.Tools))
	for i, tool := range declaration.Tools {
		tools[i] = toolDeclaration(tool)
	}
	result := &contracts.ToolProviderDeclaration{
		Identity: contracts.ToolProviderIdentity{
			Author:      declaration.Identity.Author,
			Name:        declaration.Identity.Name,
			Description: toolI18N(declaration.Identity.Description),
			Icon:        declaration.Identity.Icon,
			IconDark:    declaration.Identity.IconDark,
			Label:       toolI18N(declaration.Identity.Label),
			Tags:        &tags,
		},
		CredentialsSchema: credentials,
		Tools:             tools,
	}
	if declaration.OAuthSchema != nil {
		clientSchema := toolProviderConfigs(declaration.OAuthSchema.ClientSchema)
		credentialsSchema := toolProviderConfigs(declaration.OAuthSchema.CredentialsSchema)
		result.OauthSchema = &contracts.OAuthSchema{
			ClientSchema:      &clientSchema,
			CredentialsSchema: &credentialsSchema,
		}
	}
	return result
}

func toolDeclaration(tool plugin_entities.ToolDeclaration) contracts.ToolDeclaration {
	var parameters []contracts.ToolParameter
	if tool.Parameters != nil {
		parameters = make([]contracts.ToolParameter, len(tool.Parameters))
		for i, parameter := range tool.Parameters {
			parameters[i] = toolParameter(parameter)
		}
	}
	result := contracts.ToolDeclaration{
		Identity: contracts.ToolIdentity{
			Author: tool.Identity.Author,
			Name:   tool.Identity.Name,
			Label:  toolI18N(tool.Identity.Label),
		},
		Description: contracts.ToolDescription{
			Human: toolI18N(tool.Description.Human),
			Llm:   tool.Description.LLM,
		},
		Parameters:           &parameters,
		HasRuntimeParameters: tool.HasRuntimeParameters,
	}
	if len(tool.OutputSchema) > 0 {
		outputSchema := map[string]any(tool.OutputSchema)
		result.OutputSchema = &outputSchema
	}
	return result
}

func toolParameter(parameter plugin_entities.ToolParameter) contracts.ToolParameter {
	var options []contracts.ParameterOption
	if parameter.Options != nil {
		options = make([]contracts.ParameterOption, len(parameter.Options))
		for i, option := range parameter.Options {
			options[i] = contracts.ParameterOption{
				Value: option.Value,
				Label: toolI18N(option.Label),
				Icon:  option.Icon,
			}
		}
	}
	result := contracts.ToolParameter{
		Name:             parameter.Name,
		Label:            toolI18N(parameter.Label),
		HumanDescription: toolI18N(parameter.HumanDescription),
		Type:             contracts.ToolParameterType(parameter.Type),
		Scope:            parameter.Scope,
		Form:             contracts.ToolParameterForm(parameter.Form),
		LlmDescription:   parameter.LLMDescription,
		Required:         parameter.Required,
		Default:          parameter.Default,
		Min:              parameter.Min,
		Max:              parameter.Max,
		Multiple:         parameter.Multiple,
		Precision:        parameter.Precision,
		Options:          &options,
		ResetOnChange:    toolResetOnChange(parameter.ResetOnChange),
	}
	if parameter.AutoGenerate != nil {
		result.AutoGenerate = &contracts.ParameterAutoGenerate{
			Type: contracts.ParameterAutoGenerateType(parameter.AutoGenerate.Type),
		}
	}
	if parameter.Template != nil {
		result.Template = &contracts.ParameterTemplate{Enabled: parameter.Template.Enabled}
	}
	return result
}

func toolProviderConfigs(configs []plugin_entities.ProviderConfig) []contracts.ProviderConfig {
	if configs == nil {
		return nil
	}
	result := make([]contracts.ProviderConfig, len(configs))
	for i, config := range configs {
		var options []contracts.ConfigOption
		if config.Options != nil {
			options = make([]contracts.ConfigOption, len(config.Options))
			for j, option := range config.Options {
				options[j] = contracts.ConfigOption{Value: option.Value, Label: toolI18N(option.Label)}
			}
		}
		result[i] = contracts.ProviderConfig{
			Name:          config.Name,
			Type:          contracts.ConfigType(config.Type),
			Scope:         config.Scope,
			Required:      config.Required,
			Default:       config.Default,
			Options:       &options,
			Multiple:      config.Multiple,
			Label:         toolI18N(config.Label),
			Help:          toolNullableI18N(config.Help),
			URL:           config.URL,
			Placeholder:   toolNullableI18N(config.Placeholder),
			ResetOnChange: toolResetOnChange(config.ResetOnChange),
		}
	}
	return result
}

func toolI18N(value plugin_entities.I18nObject) contracts.I18NObject {
	return contracts.I18NObject{
		EnUS:   value.EnUS,
		JaJP:   toolOptionalString(value.JaJp),
		ZhHans: toolOptionalString(value.ZhHans),
		PtBR:   toolOptionalString(value.PtBr),
	}
}

func toolNullableI18N(value *plugin_entities.I18nObject) *contracts.I18NObject {
	if value == nil {
		return nil
	}
	result := toolI18N(*value)
	return &result
}

func toolOptionalString(value string) *string {
	if value == "" {
		return nil
	}
	return &value
}

func toolResetOnChange(value []string) *contracts.ResetOnChange {
	if len(value) == 0 {
		return nil
	}
	return &value
}

package command

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/langgenius/dify-plugin-daemon/cmd/dify_cli/config"
	"github.com/langgenius/dify-plugin-daemon/cmd/dify_cli/types"
	"github.com/langgenius/dify-plugin-daemon/pkg/entities/plugin_entities"
	"github.com/langgenius/dify-plugin-daemon/pkg/entities/requests"
	"github.com/langgenius/dify-plugin-daemon/pkg/utils/http_requests"
	"github.com/spf13/cobra"
)

var PullCmd = &cobra.Command{
	Use:   "pull",
	Short: "Pull all tools from Dify and save to config",
	Long:  `Pull all available tools from Dify platform and save to .dify_cli.json`,
	Run:   runPull,
}

type pullRequest struct {
	TenantID string `json:"tenant_id"`
	UserID   string `json:"user_id"`
}

type toolParameter struct {
	Name             string                            `json:"name"`
	Type             plugin_entities.ToolParameterType `json:"type"`
	Label            plugin_entities.I18nObject        `json:"label"`
	Required         bool                              `json:"required"`
	Default          any                               `json:"default"`
	Options          []plugin_entities.ParameterOption `json:"options"`
	HumanDescription plugin_entities.I18nObject        `json:"human_description"`
}

type toolEntity struct {
	Author      string                     `json:"author"`
	Name        string                     `json:"name"`
	Label       plugin_entities.I18nObject `json:"label"`
	Description plugin_entities.I18nObject `json:"description"`
	Parameters  []toolParameter            `json:"parameters"`
}

type providerEntity struct {
	ID                     string            `json:"id"`
	Author                 string            `json:"author"`
	Name                   string            `json:"name"`
	Description            map[string]string `json:"description"`
	Label                  map[string]string `json:"label"`
	Type                   string            `json:"type"`
	Tools                  []toolEntity      `json:"tools"`
	IsTeamAuthorization    bool              `json:"is_team_authorization"`
	PluginID               string            `json:"plugin_id"`
	PluginUniqueIdentifier string            `json:"plugin_unique_identifier"`
}

type pullResponse struct {
	Data struct {
		Providers []providerEntity `json:"providers"`
	} `json:"data"`
	Error string `json:"error"`
}

func runPull(cmd *cobra.Command, args []string) {
	envCfg := loadEnvConfig()

	providers, err := fetchProviders(envCfg)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: failed to fetch tools: %v\n", err)
		os.Exit(1)
	}

	tools := convertProvidersToTools(providers)

	cfg := &types.DifyConfig{
		Env:   envCfg,
		Tools: tools,
	}

	if err := saveConfig(cfg); err != nil {
		fmt.Fprintf(os.Stderr, "Error: failed to save config: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("Successfully pulled %d tools from %d providers\n", len(tools), len(providers))
	fmt.Printf("Config saved to %s\n", config.GetConfigPath())
}

func loadEnvConfig() types.EnvConfig {
	existingCfg, err := config.Load()
	if err == nil {
		return existingCfg.Env
	}

	innerAPIURL := os.Getenv("DIFY_INNER_API_URL")
	innerAPIKey := os.Getenv("DIFY_INNER_API_KEY")
	tenantID := os.Getenv("DIFY_TENANT_ID")
	userID := os.Getenv("DIFY_USER_ID")

	if innerAPIURL == "" || innerAPIKey == "" {
		fmt.Fprintln(os.Stderr, "Error: DIFY_INNER_API_URL and DIFY_INNER_API_KEY environment variables are required")
		fmt.Fprintln(os.Stderr, "Or create a .dify_cli.json config file first")
		os.Exit(1)
	}

	return types.EnvConfig{
		InnerAPIURL: innerAPIURL,
		InnerAPIKey: innerAPIKey,
		TenantID:    tenantID,
		UserID:      userID,
	}
}

func fetchProviders(envCfg types.EnvConfig) ([]providerEntity, error) {
	reqBody := pullRequest{
		TenantID: envCfg.TenantID,
		UserID:   envCfg.UserID,
	}

	url := strings.TrimSuffix(envCfg.InnerAPIURL, "/") + "/inner/api/fetch/tools/list"

	client := &http.Client{
		Timeout: 2 * time.Minute,
	}

	resp, err := http_requests.Request(
		client,
		url,
		"POST",
		http_requests.HttpHeader(map[string]string{
			"X-Inner-Api-Key": envCfg.InnerAPIKey,
			"Content-Type":    "application/json",
		}),
		http_requests.HttpPayloadJson(reqBody),
		http_requests.HttpWriteTimeout(5000),
		http_requests.HttpReadTimeout(120000),
	)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("API returned status %d", resp.StatusCode)
	}

	var result pullResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	if result.Error != "" {
		return nil, fmt.Errorf("API error: %s", result.Error)
	}

	return result.Data.Providers, nil
}

func convertProvidersToTools(providers []providerEntity) []types.DifyToolDeclaration {
	var tools []types.DifyToolDeclaration

	for _, provider := range providers {
		if !provider.IsTeamAuthorization {
			continue
		}

		providerType := mapProviderType(provider.Type)

		for _, tool := range provider.Tools {
			params := make([]plugin_entities.ToolParameter, 0, len(tool.Parameters))
			for _, p := range tool.Parameters {
				params = append(params, plugin_entities.ToolParameter{
					Name:             p.Name,
					Type:             p.Type,
					Label:            p.Label,
					Required:         p.Required,
					Default:          p.Default,
					Options:          p.Options,
					HumanDescription: p.HumanDescription,
				})
			}

			toolDecl := types.DifyToolDeclaration{
				ProviderType: providerType,
				Identity: types.DifyToolIdentity{
					Author:   tool.Author,
					Name:     tool.Name,
					Label:    tool.Label,
					Provider: provider.Name,
				},
				Description: plugin_entities.ToolDescription{
					Human: tool.Description,
					LLM:   getDescriptionText(tool.Description),
				},
				Parameters:     params,
				CredentialId:   provider.ID,
				CredentialType: "default",
			}

			tools = append(tools, toolDecl)
		}
	}

	return tools
}

func mapProviderType(providerType string) requests.ToolType {
	switch providerType {
	case "builtin":
		return requests.TOOL_TYPE_BUILTIN
	case "api":
		return requests.TOOL_TYPE_API
	case "workflow":
		return requests.TOOL_TYPE_WORKFLOW
	case "mcp":
		return requests.TOOL_TYPE_MCP
	default:
		return requests.TOOL_TYPE_BUILTIN
	}
}

func getDescriptionText(desc plugin_entities.I18nObject) string {
	if desc.EnUS != "" {
		return desc.EnUS
	}
	if desc.ZhHans != "" {
		return desc.ZhHans
	}
	return ""
}

func saveConfig(cfg *types.DifyConfig) error {
	data, err := json.MarshalIndent(cfg, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(config.GetConfigPath(), data, 0644)
}

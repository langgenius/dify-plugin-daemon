package config

import (
	"bufio"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"

	"github.com/langgenius/dify-plugin-daemon/cmd/dify_cli/types"
	"github.com/langgenius/dify-plugin-daemon/pkg/entities/plugin_entities"
)

const configFileName = ".dify_cli.json"

func GetConfigPath() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(home, configFileName), nil
}

func LoadEnvFile(path string) (types.EnvConfig, error) {
	var config types.EnvConfig

	file, err := os.Open(path)
	if err != nil {
		return config, err
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		parts := strings.SplitN(line, "=", 2)
		if len(parts) != 2 {
			continue
		}

		key := strings.TrimSpace(parts[0])
		value := strings.Trim(strings.TrimSpace(parts[1]), "\"'")

		switch key {
		case "INNER_API_URL":
			config.InnerAPIURL = value
		case "INNER_API_KEY":
			config.InnerAPIKey = value
		}
	}

	return config, scanner.Err()
}

func LoadSchemaFile(path string) ([]plugin_entities.ToolProviderDeclaration, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	var providers []plugin_entities.ToolProviderDeclaration
	if err := json.Unmarshal(data, &providers); err != nil {
		return nil, err
	}

	return providers, nil
}

func Save(config *types.DifyConfig) error {
	path, err := GetConfigPath()
	if err != nil {
		return err
	}

	data, err := json.MarshalIndent(config, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(path, data, 0600)
}

func Load() (*types.DifyConfig, error) {
	path, err := GetConfigPath()
	if err != nil {
		return nil, err
	}

	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	var config types.DifyConfig
	if err := json.Unmarshal(data, &config); err != nil {
		return nil, err
	}

	return &config, nil
}

func FindTool(config *types.DifyConfig, toolName string) (*plugin_entities.ToolProviderDeclaration, *plugin_entities.ToolDeclaration) {
	for i := range config.Providers {
		provider := &config.Providers[i]
		for j := range provider.Tools {
			tool := &provider.Tools[j]
			if tool.Identity.Name == toolName {
				return provider, tool
			}
		}
	}
	return nil, nil
}

func GetSelfPath() (string, error) {
	exe, err := os.Executable()
	if err != nil {
		return "", err
	}
	return filepath.EvalSymlinks(exe)
}

func GetBinDir() string {
	home, err := os.UserHomeDir()
	if err != nil {
		return "/usr/local/bin"
	}
	return filepath.Join(home, ".dify", "bin")
}

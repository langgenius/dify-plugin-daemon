package config

import (
	"bufio"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"

	"github.com/langgenius/dify-plugin-daemon/cmd/dify_cli/types"
)

const configFileName = ".dify_cli.json"

func GetConfigPath() string {
	return configFileName
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
		case "TENANT_ID":
			config.TenantID = value
		case "USER_ID":
			config.UserID = value
		}
	}
	return config, scanner.Err()
}

func LoadSchemaFile(path string) (*types.ToolSchemas, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var schemas types.ToolSchemas
	if err := json.Unmarshal(data, &schemas); err != nil {
		return nil, err
	}
	return &schemas, nil
}

func Save(config *types.DifyConfig) error {
	data, err := json.MarshalIndent(config, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(GetConfigPath(), data, 0600)
}

func Load() (*types.DifyConfig, error) {
	data, err := os.ReadFile(GetConfigPath())
	if err != nil {
		return nil, err
	}
	var config types.DifyConfig
	if err := json.Unmarshal(data, &config); err != nil {
		return nil, err
	}
	return &config, nil
}

func FindTool(config *types.DifyConfig, toolName string) *types.DifyToolDeclaration {
	for i := range config.Tools {
		if config.Tools[i].Identity.Name == toolName {
			return &config.Tools[i]
		}
	}
	return nil
}

func GetSelfPath() (string, error) {
	exe, err := os.Executable()
	if err != nil {
		return "", err
	}
	return filepath.EvalSymlinks(exe)
}

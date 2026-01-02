package config

import (
	"encoding/json"
	"os"
	"path/filepath"

	"github.com/langgenius/dify-plugin-daemon/cmd/dify_cli/types"
)

const configFileName = ".dify_cli.json"

func GetConfigPath() string {
	return configFileName
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
		tool := &config.Tools[i]
		if tool.Identity.Name == toolName {
			return tool
		}
		if tool.Identity.Provider+"/"+tool.Identity.Name == toolName {
			return tool
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

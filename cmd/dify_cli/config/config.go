package config

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"

	"github.com/langgenius/dify-plugin-daemon/cmd/dify_cli/types"
)

const configFileName = ".dify_cli.json"

func GetConfigPath() string {
	if envPath := os.Getenv("DIFY_CLI_CONFIG"); envPath != "" {
		return envPath
	}
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

func FindToolReference(config *types.DifyConfig, name string) *types.ToolReference {
	for i := range config.ToolReferences {
		ref := &config.ToolReferences[i]
		refName := ref.ToolName + "_" + ref.ID
		if refName == name {
			return ref
		}
	}
	return nil
}

func FindToolByReference(config *types.DifyConfig, ref *types.ToolReference) *types.DifyToolDeclaration {
	for i := range config.Tools {
		tool := &config.Tools[i]
		if tool.Identity.Name == ref.ToolName && tool.Identity.Provider == ref.ToolProvider {
			return tool
		}
	}
	return nil
}

func GetReferenceSymlinkName(ref *types.ToolReference) string {
	return ref.ToolName + "_" + ref.ID
}

func ParseSymlinkName(name string) (toolName string, refID string, isReference bool) {
	idx := strings.LastIndex(name, "_")
	if idx == -1 {
		return name, "", false
	}
	potentialID := name[idx+1:]
	potentialToolName := name[:idx]
	if len(potentialID) > 0 && len(potentialToolName) > 0 {
		return potentialToolName, potentialID, true
	}
	return name, "", false
}

func GetSelfPath() (string, error) {
	exe, err := os.Executable()
	if err != nil {
		return "", err
	}
	return filepath.EvalSymlinks(exe)
}

package local_runtime

import (
	_ "embed"
	"fmt"
	"os"
	"os/exec"
	"path"
	"regexp"
	"strings"

	version "github.com/hashicorp/go-version"
	"github.com/langgenius/dify-plugin-daemon/pkg/utils/log"
)

//go:embed patches/0.0.1b70.ai_model.py.patch
var python001b70aiModelsPatches []byte

//go:embed patches/0.1.1.llm.py.patch
var python011llmPatches []byte

//go:embed patches/0.1.1.request_reader.py.patch
var python011requestReaderPatches []byte

func (p *LocalPluginRuntime) patchPluginSdk(
	requirementsPath string,
	pythonInterpreterPath string,
) error {
	// get the version of the plugin sdk
	requirements, err := os.ReadFile(requirementsPath)
	if err != nil {
		return fmt.Errorf("failed to read requirements.txt: %s", err)
	}

	pluginSdkVersion, err := p.getPluginSdkVersion(string(requirements))
	if err != nil {
		log.Error("failed to get the version of the plugin sdk: %s", err)
		return nil
	}

	pluginSdkVersionObj, err := version.NewVersion(pluginSdkVersion)
	if err != nil {
		log.Error("failed to create the version: %s", err)
		return nil
	}

	if pluginSdkVersionObj.LessThan(version.Must(version.NewVersion("0.0.1b70"))) {
		// get dify-plugin path
		command := exec.Command(pythonInterpreterPath, "-c", "import importlib.util;print(importlib.util.find_spec('dify_plugin').origin)")
		command.Dir = p.State.WorkingPath
		output, err := command.Output()
		if err != nil {
			return fmt.Errorf("failed to get the path of the plugin sdk: %s", err)
		}

		pluginSdkPath := path.Dir(strings.TrimSpace(string(output)))
		patchPath := path.Join(pluginSdkPath, "interfaces/model/ai_model.py")

		// apply the patch
		if _, err := os.Stat(patchPath); err != nil {
			return fmt.Errorf("failed to find the patch file: %s", err)
		}

		if err := os.WriteFile(patchPath, python001b70aiModelsPatches, 0644); err != nil {
			return fmt.Errorf("failed to write the patch file: %s", err)
		}
	}

	if pluginSdkVersionObj.LessThan(version.Must(version.NewVersion("0.1.1"))) {
		// get dify-plugin path
		command := exec.Command(pythonInterpreterPath, "-c", "import importlib.util;print(importlib.util.find_spec('dify_plugin').origin)")
		command.Dir = p.State.WorkingPath
		output, err := command.Output()
		if err != nil {
			return fmt.Errorf("failed to get the path of the plugin sdk: %s", err)
		}

		pluginSdkPath := path.Dir(strings.TrimSpace(string(output)))
		patchPath := path.Join(pluginSdkPath, "entities/model/llm.py")

		// apply the patch
		if _, err := os.Stat(patchPath); err != nil {
			return fmt.Errorf("failed to find the patch file: %s", err)
		}

		if err := os.WriteFile(patchPath, python011llmPatches, 0644); err != nil {
			return fmt.Errorf("failed to write the patch file: %s", err)
		}

		patchPath = path.Join(pluginSdkPath, "core/server/stdio/request_reader.py")
		if _, err := os.Stat(patchPath); err != nil {
			return fmt.Errorf("failed to find the patch file: %s", err)
		}

		if err := os.WriteFile(patchPath, python011requestReaderPatches, 0644); err != nil {
			return fmt.Errorf("failed to write the patch file: %s", err)
		}
	}
	return nil
}

func (p *LocalPluginRuntime) getPluginSdkVersion(requirements string) (string, error) {
	// using regex to find the version of the plugin sdk
	// First try to match exact version or compatible version
	re := regexp.MustCompile(`(?:dify[_-]plugin)(?:~=|==)([0-9.a-z]+)`)
	matches := re.FindStringSubmatch(requirements)
	if len(matches) >= 2 {
		return matches[1], nil
	}

	// Try to match version ranges with multiple constraints
	// Extract all version constraints for dify-plugin
	// Try to match version ranges with multiple constraints
	// For example: dify-plugin>=0.1.0,<0.2.0
	reAllConstraints := regexp.MustCompile(`(?:dify[_-]plugin)([><]=?|==)([0-9.a-z]+)(?:,([><]=?|==)([0-9.a-z]+))?`)
	allMatches := reAllConstraints.FindAllStringSubmatch(requirements, -1)

	if len(allMatches) > 0 {
		// Always return the highest version among all constraints
		var highestVersion *version.Version
		var versionStr string

		for _, match := range allMatches {
			// Check for the second version constraint if it exists
			if len(match) >= 5 {
				currentVersionStr := match[4]
				currentVersion, err := version.NewVersion(currentVersionStr)
				if err != nil {
					continue
				}

				if highestVersion == nil || currentVersion.GreaterThan(highestVersion) {
					highestVersion = currentVersion
					versionStr = currentVersionStr
				}
			} else if len(match) >= 3 {
				currentVersionStr := match[2]
				currentVersion, err := version.NewVersion(currentVersionStr)
				if err != nil {
					continue
				}

				if highestVersion == nil || currentVersion.GreaterThan(highestVersion) {
					highestVersion = currentVersion
					versionStr = currentVersionStr
				}
			}
		}

		if versionStr != "" {
			return versionStr, nil
		}

		// If we couldn't parse any versions but have matches, return the first one
		if len(allMatches[0]) >= 3 {
			return allMatches[0][2], nil
		}
	}

	return "", fmt.Errorf("failed to find the version of the plugin sdk")
}

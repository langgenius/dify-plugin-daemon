package local_runtime

import (
	_ "embed"
	"fmt"

	"github.com/langgenius/dify-plugin-daemon/pkg/utils/log"
)

func (p *LocalPluginRuntime) InitPythonEnvironment() error {
	// prepare uv environment
	uvPath, err := p.prepareUV()
	if err != nil {
		return fmt.Errorf("failed to find uv path: %w", err)
	}

	// check if virtual environment exists
	venv, err := p.checkPythonVirtualEnvironment()
	switch err {
	case ErrVirtualEnvironmentInvalid:
		// remove the venv and rebuild it
		log.Warn("virtual environment for %s is invalid; deleting and recreating", p.Config.Identity())
		p.deleteVirtualEnvironment()

		// create virtual environment
		venv, err = p.createVirtualEnvironment(uvPath)
		if err != nil {
			return fmt.Errorf("failed to create virtual environment: %w", err)
		}
	case ErrVirtualEnvironmentNotFound:
		// create virtual environment
		venv, err = p.createVirtualEnvironment(uvPath)
		if err != nil {
			return fmt.Errorf("failed to create virtual environment: %w", err)
		}
	case nil:
		// PATCH:
		//  plugin sdk version less than 0.0.1b70 contains a memory leak bug
		//  to reach a better user experience, we will patch it here using a patched file
		// https://github.com/langgenius/dify-plugin-sdks/commit/161045b65f708d8ef0837da24440ab3872821b3b
		dependencyFilePath, err := p.getDependencyFilePath()
		if err != nil {
			log.Error("failed to get dependency file path for patching", "error", err)
		} else if err := p.patchPluginSdk(
			dependencyFilePath,
			venv.pythonInterpreterPath,
		); err != nil {
			log.Error("failed to patch the plugin sdk", "error", err)
		}

		// everything is good, return nil
		return nil
	default:
		return fmt.Errorf("failed to check virtual environment: %w", err)
	}

	// detect dependency file type and install dependencies
	dependencyFileType, err := p.detectDependencyFileType()
	if err != nil {
		return fmt.Errorf("failed to detect dependency file: %w", err)
	}

	if err := p.installDependencies(uvPath, dependencyFileType); err != nil {
		return fmt.Errorf("failed to install dependencies: %w", err)
	}

	// pre-compile the plugin to avoid costly compilation on first invocation
	if err := p.preCompile(venv.pythonInterpreterPath); err != nil {
		return fmt.Errorf("failed to pre-compile the plugin: %w", err)
	}

	// PATCH:
	//  plugin sdk version less than 0.0.1b70 contains a memory leak bug
	//  to reach a better user experience, we will patch it here using a patched file
	// https://github.com/langgenius/dify-plugin-sdks/commit/161045b65f708d8ef0837da24440ab3872821b3b
	dependencyFilePath, err := p.getDependencyFilePath()
	if err != nil {
		log.Error("failed to get dependency file path for patching", "error", err)
	} else if err := p.patchPluginSdk(
		dependencyFilePath,
		venv.pythonInterpreterPath,
	); err != nil {
		log.Error("failed to patch the plugin sdk", "error", err)
	}

	// mark the virtual environment as valid if everything goes well
	if err := p.markVirtualEnvironmentAsValid(); err != nil {
		log.Error("failed to mark the virtual environment as valid", "error", err)
	}

	return nil
}

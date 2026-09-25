package controlpanel

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/langgenius/dify-plugin-daemon/pkg/entities/plugin_entities"
)

// InstallToLocal installs a plugin to the local plugin runtime
// It's scope only for marking the plugin as `installed`,
// you should call `LaunchLocalPlugin` to start it or it may launched by daemon
// automatically
func (c *ControlPanel) InstallToLocal(
	pluginUniqueIdentifier plugin_entities.PluginUniqueIdentifier,
) error {
	// copy the package from `packageBucket` to `installedBucket`
	// this step marks the plugin as `installed`
	packageFile, err := c.packageBucket.Get(pluginUniqueIdentifier.String())
	if err != nil {
		return errors.Join(
			errors.New("failed to get package file when trying to install plugin to local"),
			err,
		)
	}

	err = c.installedBucket.Save(pluginUniqueIdentifier, packageFile)
	if err != nil {
		return errors.Join(
			errors.New("failed to save package file to installed bucket when trying to install plugin to local"),
			err,
		)
	}

	// try to decode the package
	decoder, _, err := c.buildPluginDecoder(pluginUniqueIdentifier)
	if err != nil {
		return err
	}

	_, err = decoder.Manifest()
	if err != nil {
		return errors.Join(
			errors.New("failed to get manifest when trying to install plugin to local"),
			err,
		)
	}

	return nil
}

// RemoveLocalPlugin removes a plugin from the local plugin runtime
// It's scope only for marking the plugin as `not installed`
// If you want to stop plugin runtime immediately, you should call `ShutdownLocalPluginForcefully`
// or `ShutdownLocalPluginGracefully`
// they have the right to shutdown a runtime.
func (c *ControlPanel) RemoveLocalPlugin(
	pluginUniqueIdentifier plugin_entities.PluginUniqueIdentifier,
) error {
	// remove the package from the `installedBucket`
	err := c.installedBucket.Delete(pluginUniqueIdentifier)
	if err != nil && !os.IsNotExist(err) {
		return errors.Join(
			errors.New("failed to delete package file from installed bucket when trying to remove plugin from local"),
			err,
		)
	}

	return nil
}

func (c *ControlPanel) RemoveLocalPluginStorage(
	pluginUniqueIdentifier plugin_entities.PluginUniqueIdentifier,
) error {
	var errs []error

	if err := c.packageBucket.Delete(pluginUniqueIdentifier.String()); err != nil && !os.IsNotExist(err) {
		errs = append(errs, errors.Join(
			errors.New("failed to delete package file from package bucket when trying to remove plugin from local"),
			err,
		))
	}

	workingPath, err := c.localPluginWorkingPath(pluginUniqueIdentifier)
	if err != nil {
		errs = append(errs, err)
	} else if err := os.RemoveAll(workingPath); err != nil {
		errs = append(errs, errors.Join(
			fmt.Errorf("failed to delete plugin working directory %s", workingPath),
			err,
		))
	}

	return errors.Join(errs...)
}

func (c *ControlPanel) localPluginWorkingPath(
	pluginUniqueIdentifier plugin_entities.PluginUniqueIdentifier,
) (string, error) {
	identity, _, ok := strings.Cut(pluginUniqueIdentifier.String(), "@")
	if !ok {
		return "", fmt.Errorf("invalid plugin unique identifier: %s", pluginUniqueIdentifier.String())
	}
	identity = strings.ReplaceAll(identity, ":", "-")

	base, err := filepath.Abs(c.config.PluginWorkingPath)
	if err != nil {
		return "", err
	}
	target, err := filepath.Abs(filepath.Join(base, fmt.Sprintf("%s@%s", identity, pluginUniqueIdentifier.Checksum())))
	if err != nil {
		return "", err
	}

	if target == base || !strings.HasPrefix(target, base+string(os.PathSeparator)) {
		return "", fmt.Errorf("refusing to delete plugin working directory outside base path: %s", target)
	}

	return target, nil
}

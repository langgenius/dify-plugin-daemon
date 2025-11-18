package controlpanel

import (
	"errors"

	"github.com/langgenius/dify-plugin-daemon/pkg/entities/plugin_entities"
)

// InstallToLocalFromPkg installs a plugin to the local plugin runtime
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
	if err != nil {
		return errors.Join(
			errors.New("failed to delete package file from installed bucket when trying to remove plugin from local"),
			err,
		)
	}

	return nil
}

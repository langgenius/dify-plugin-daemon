package controlpanel

import (
	serverless "github.com/langgenius/dify-plugin-daemon/internal/core/serverless_connector"
	"github.com/langgenius/dify-plugin-daemon/pkg/entities/plugin_entities"
	"github.com/langgenius/dify-plugin-daemon/pkg/utils/stream"
)

func (c *ControlPanel) InstallToServerless(
	pluginUniqueIdentifier plugin_entities.PluginUniqueIdentifier,
) (
	*stream.Stream[serverless.LaunchFunctionResponse], error,
) {
	decoder, packageFile, err := c.buildPluginDecoder(pluginUniqueIdentifier)
	if err != nil {
		return nil, err
	}

	// check valid manifest
	_, err = decoder.Manifest()
	if err != nil {
		return nil, err
	}

	// serverless.LaunchPlugin will check if the plugin has already been launched, if so, it returns directly
	response, err := serverless.LaunchPlugin(
		pluginUniqueIdentifier,
		packageFile,
		decoder,
		c.config.DifyPluginServerlessConnectorLaunchTimeout,
		false,
	)
	if err != nil {
		return nil, err
	}

	return response, nil
}

func (c *ControlPanel) ReinstallToServerless(
	pluginUniqueIdentifier plugin_entities.PluginUniqueIdentifier,
) (
	*stream.Stream[serverless.LaunchFunctionResponse], error,
) {
	decoder, packageFile, err := c.buildPluginDecoder(pluginUniqueIdentifier)
	if err != nil {
		return nil, err
	}

	// check valid manifest
	_, err = decoder.Manifest()
	if err != nil {
		return nil, err
	}

	response, err := serverless.LaunchPlugin(
		pluginUniqueIdentifier,
		packageFile,
		decoder,
		c.config.DifyPluginServerlessConnectorLaunchTimeout,
		true, // ignoreIdempotent, true means always reinstall
	)
	if err != nil {
		return nil, err
	}

	return response, nil
}

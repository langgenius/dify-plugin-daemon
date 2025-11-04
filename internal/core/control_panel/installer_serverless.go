package controlpanel

import (
	serverless "github.com/langgenius/dify-plugin-daemon/internal/core/plugin_manager/serverless_connector"
	"github.com/langgenius/dify-plugin-daemon/internal/utils/stream"
	"github.com/langgenius/dify-plugin-daemon/pkg/plugin_packager/decoder"
)

func (c *ControlPanel) InstallToServerlessFromPkg(
	packageFile []byte,
	decoder decoder.PluginDecoder,
) (
	*stream.Stream[InstallServerlessPluginResponse], error,
) {
	// check valid manifest
	_, err := decoder.Manifest()
	if err != nil {
		return nil, err
	}

	uniqueIdentity, err := decoder.UniqueIdentity()
	if err != nil {
		return nil, err
	}

	// serverless.LaunchPlugin will check if the plugin has already been launched, if so, it returns directly
	response, err := serverless.LaunchPlugin(
		uniqueIdentity,
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

func (c *ControlPanel) ReinstallToServerlessFromPkg(
	originalPackager []byte,
	decoder decoder.PluginDecoder,
) (
	*stream.Stream[InstallServerlessPluginResponse], error,
) {
	_, err := decoder.Manifest()
	if err != nil {
		return nil, err
	}
	uniqueIdentifier, err := decoder.UniqueIdentity()
	if err != nil {
		return nil, err
	}

	response, err := serverless.LaunchPlugin(
		uniqueIdentifier,
		originalPackager,
		decoder,
		c.config.DifyPluginServerlessConnectorLaunchTimeout,
		true, // ignoreIdempotent, true means always reinstall
	)
	if err != nil {
		return nil, err
	}

	return response, nil
}

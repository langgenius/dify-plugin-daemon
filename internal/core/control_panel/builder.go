package controlpanel

import (
	"errors"
	"fmt"

	"github.com/langgenius/dify-plugin-daemon/internal/core/local_runtime"
	"github.com/langgenius/dify-plugin-daemon/pkg/entities/plugin_entities"
	"github.com/langgenius/dify-plugin-daemon/pkg/plugin_packager/decoder"
)

// buildLocalPluginRuntime builds a local plugin runtime and returns the runtime and the decoder
func (c *ControlPanel) buildLocalPluginRuntime(
	pluginUniqueIdentifier plugin_entities.PluginUniqueIdentifier,
) (*local_runtime.LocalPluginRuntime, *decoder.ZipPluginDecoder, error) {
	decoder, _, err := c.buildPluginDecoder(pluginUniqueIdentifier)
	if err != nil {
		return nil, nil, err
	}

	runtime, err := local_runtime.ConstructPluginRuntime(c.config, decoder)
	if err != nil {
		return nil, nil, errors.Join(err, fmt.Errorf("construct plugin runtime error"))
	}

	return runtime, decoder, nil
}

// buildPluginDecoder builds a plugin decoder and returns the decoder and the plugin zip file
func (c *ControlPanel) buildPluginDecoder(
	pluginUniqueIdentifier plugin_entities.PluginUniqueIdentifier,
) (*decoder.ZipPluginDecoder, []byte, error) {
	pluginZip, err := c.packageBucket.Get(pluginUniqueIdentifier.String())
	if err != nil {
		return nil, nil, errors.Join(err, fmt.Errorf("get plugin package error"))
	}

	decoder, err := decoder.NewZipPluginDecoderWithThirdPartySignatureVerificationConfig(
		pluginZip, &decoder.ThirdPartySignatureVerificationConfig{
			Enabled:        c.config.ThirdPartySignatureVerificationEnabled,
			PublicKeyPaths: c.config.ThirdPartySignatureVerificationPublicKeys,
		},
	)
	if err != nil {
		return nil, nil, errors.Join(err, fmt.Errorf("create plugin decoder error"))
	}

	return decoder, pluginZip, nil
}

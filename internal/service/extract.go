package service

import (
	"github.com/langgenius/dify-plugin-daemon/internal/core/plugin_manager"
	"github.com/langgenius/dify-plugin-daemon/internal/types/app"
	"github.com/langgenius/dify-plugin-daemon/internal/types/exception"
	"github.com/langgenius/dify-plugin-daemon/pkg/entities"
	"github.com/langgenius/dify-plugin-daemon/pkg/entities/plugin_entities"
	"github.com/langgenius/dify-plugin-daemon/pkg/plugin_packager/decoder"
)

func ExtractPluginSchema(
	config *app.Config,
	pluginUniqueIdentifier plugin_entities.PluginUniqueIdentifier,
) *entities.Response {
	manager := plugin_manager.Manager()
	pkgFile, err := manager.GetPackage(pluginUniqueIdentifier)
	if err != nil {
		return exception.BadRequestError(err).ToResponse()
	}

	zipDecoder, err := decoder.NewZipPluginDecoderWithThirdPartySignatureVerificationConfig(
		pkgFile,
		&decoder.ThirdPartySignatureVerificationConfig{
			Enabled:        config.ThirdPartySignatureVerificationEnabled,
			PublicKeyPaths: config.ThirdPartySignatureVerificationPublicKeys,
		},
	)
	if err != nil {
		return exception.BadRequestError(err).ToResponse()
	}
	defer zipDecoder.Close()

	verification, _ := zipDecoder.Verification()
	if verification == nil && zipDecoder.Verified() {
		verification = decoder.DefaultVerification()
	}

	declaration, err := zipDecoder.Manifest()
	if err != nil {
		return exception.BadRequestError(err).ToResponse()
	}

	return entities.NewSuccessResponse(map[string]any{
		"unique_identifier": pluginUniqueIdentifier,
		"manifest":          declaration,
		"verification":      verification,
	})
}

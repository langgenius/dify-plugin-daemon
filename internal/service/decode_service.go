package service

import (
	"github.com/langgenius/dify-plugin-daemon/internal/core/plugin_manager"
	"github.com/langgenius/dify-plugin-daemon/internal/db"
	"github.com/langgenius/dify-plugin-daemon/internal/types/app"
	"github.com/langgenius/dify-plugin-daemon/internal/types/exception"
	"github.com/langgenius/dify-plugin-daemon/internal/types/models"
	"github.com/langgenius/dify-plugin-daemon/pkg/entities"
	"github.com/langgenius/dify-plugin-daemon/pkg/entities/plugin_entities"
	"github.com/langgenius/dify-plugin-daemon/pkg/plugin_packager/decoder"
)

func FetchPluginFromIdentifier(
	pluginUniqueIdentifier plugin_entities.PluginUniqueIdentifier,
) *entities.Response {
	_, err := db.GetOne[models.Plugin](
		db.Equal("plugin_unique_identifier", pluginUniqueIdentifier.String()),
	)
	if err == db.ErrDatabaseNotFound {
		return entities.NewSuccessResponse(false)
	}
	if err != nil {
		return exception.InternalServerError(err).ToResponse()
	}

	return entities.NewSuccessResponse(true)
}

/*
 * Decode a plugin from a given identifier, no tenant_id is needed
 * When upload local plugin inside Dify, the second step need to ensure that the plugin is valid
 * So we need to provide a way to decode the plugin and verify the signature
 */
func DecodePluginFromIdentifier(
	config *app.Config,
	pluginUniqueIdentifier plugin_entities.PluginUniqueIdentifier,
) *entities.Response {
	// get plugin package and decode again
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

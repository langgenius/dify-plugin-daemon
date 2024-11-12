package service

import (
	"errors"
	"io"
	"mime/multipart"

	"github.com/gin-gonic/gin"
	"github.com/langgenius/dify-plugin-daemon/internal/core/plugin_manager"
	"github.com/langgenius/dify-plugin-daemon/internal/core/plugin_packager/decoder"
	"github.com/langgenius/dify-plugin-daemon/internal/types/app"
	"github.com/langgenius/dify-plugin-daemon/internal/types/entities"
	"github.com/langgenius/dify-plugin-daemon/internal/types/entities/plugin_entities"
	"github.com/langgenius/dify-plugin-daemon/internal/utils/cache"
)

func UploadPluginFromPkg(
	config *app.Config,
	c *gin.Context,
	tenant_id string,
	dify_pkg_file multipart.File,
	verify_signature bool,
) *entities.Response {
	pluginFile, err := io.ReadAll(dify_pkg_file)
	if err != nil {
		return entities.NewErrorResponse(-500, err.Error())
	}

	decoder, err := decoder.NewZipPluginDecoder(pluginFile)
	if err != nil {
		return entities.NewErrorResponse(-500, err.Error())
	}

	pluginUniqueIdentifier, err := decoder.UniqueIdentity()
	if err != nil {
		return entities.NewErrorResponse(-500, err.Error())
	}

	manager := plugin_manager.Manager()
	declaration, err := manager.SavePackage(pluginUniqueIdentifier, pluginFile)
	if err != nil {
		return entities.NewErrorResponse(-500, err.Error())
	}

	if config.ForceVerifyingSignature || verify_signature {
		if !declaration.Verified {
			return entities.NewErrorResponse(-500, errors.Join(err, errors.New(
				"plugin verification has been enabled, and the plugin you want to install has a bad signature",
			)).Error())
		}
	}

	return entities.NewSuccessResponse(map[string]any{
		"unique_identifier": pluginUniqueIdentifier,
		"manifest":          declaration,
	})
}

func FetchPluginManifest(
	tenant_id string,
	pluginUniqueIdentifier plugin_entities.PluginUniqueIdentifier,
) *entities.Response {
	type ManifestCache struct {
		Declaration plugin_entities.PluginDeclaration `json:"declaration"`
	}

	pluginManifestCache, err := cache.AutoGetWithGetter(pluginUniqueIdentifier.String(), func() (*ManifestCache, error) {
		manager := plugin_manager.Manager()
		pkg, err := manager.GetPackage(pluginUniqueIdentifier)
		if err != nil {
			return nil, err
		}

		decoder, err := decoder.NewZipPluginDecoder(pkg)
		if err != nil {
			return nil, err
		}

		manifest, err := decoder.Manifest()
		if err != nil {
			return nil, err
		}

		return &ManifestCache{Declaration: manifest}, nil
	})

	if err != nil {
		return entities.NewErrorResponse(-500, err.Error())
	}

	return entities.NewSuccessResponse(pluginManifestCache)
}

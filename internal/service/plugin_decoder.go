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
	plugin_file, err := io.ReadAll(dify_pkg_file)
	if err != nil {
		return entities.NewErrorResponse(-500, err.Error())
	}

	decoder, err := decoder.NewZipPluginDecoder(plugin_file)
	if err != nil {
		return entities.NewErrorResponse(-500, err.Error())
	}

	manifest, err := decoder.Manifest()
	if err != nil {
		return entities.NewErrorResponse(-500, err.Error())
	}

	if config.ForceVerifyingSignature || verify_signature {
		if !manifest.Verified {
			return entities.NewErrorResponse(-500, errors.Join(err, errors.New(
				"plugin verification has been enabled, and the plugin you want to install has a bad signature",
			)).Error())
		}
	}

	plugin_unique_identifier, err := decoder.UniqueIdentity()
	if err != nil {
		return entities.NewErrorResponse(-500, err.Error())
	}

	manager := plugin_manager.Manager()
	if err := manager.SavePackage(plugin_unique_identifier, plugin_file); err != nil {
		return entities.NewErrorResponse(-500, err.Error())
	}

	return entities.NewSuccessResponse(plugin_unique_identifier)
}

func FetchPluginManifest(
	tenant_id string,
	plugin_unique_identifier plugin_entities.PluginUniqueIdentifier,
) *entities.Response {
	type ManifestCache struct {
		Declaration plugin_entities.PluginDeclaration `json:"declaration"`
	}

	plugin_manifest_cache, err := cache.AutoGetWithGetter(plugin_unique_identifier.String(), func() (*ManifestCache, error) {
		manager := plugin_manager.Manager()
		pkg, err := manager.GetPackage(plugin_unique_identifier)
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

	return entities.NewSuccessResponse(plugin_manifest_cache)
}

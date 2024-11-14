package service

import (
	"errors"
	"io"
	"mime/multipart"

	"github.com/gin-gonic/gin"
	"github.com/langgenius/dify-plugin-daemon/internal/core/bundle_packager"
	"github.com/langgenius/dify-plugin-daemon/internal/core/plugin_manager"
	"github.com/langgenius/dify-plugin-daemon/internal/core/plugin_packager/decoder"
	"github.com/langgenius/dify-plugin-daemon/internal/types/app"
	"github.com/langgenius/dify-plugin-daemon/internal/types/entities"
	"github.com/langgenius/dify-plugin-daemon/internal/types/entities/bundle_entities"
	"github.com/langgenius/dify-plugin-daemon/internal/types/entities/plugin_entities"
	"github.com/langgenius/dify-plugin-daemon/internal/utils/cache"
)

func UploadPluginPkg(
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

func UploadPluginBundle(
	config *app.Config,
	c *gin.Context,
	tenant_id string,
	dify_bundle_file *multipart.FileHeader,
	verify_signature bool,
) *entities.Response {
	packager, err := bundle_packager.NewZipBundlePackager(dify_bundle_file.Filename)
	if err != nil {
		return entities.NewErrorResponse(-500, errors.Join(err, errors.New("failed to create bundle packager")).Error())
	}

	// load bundle
	bundle, err := packager.Manifest()
	if err != nil {
		return entities.NewErrorResponse(-500, errors.Join(err, errors.New("failed to load bundle manifest")).Error())
	}

	manager := plugin_manager.Manager()

	result := []map[string]any{}

	for _, dependency := range bundle.Dependencies {
		if dependency.Type == bundle_entities.DEPENDENCY_TYPE_GITHUB {
			if dep, ok := dependency.Value.(bundle_entities.GithubDependency); ok {
				result = append(result, map[string]any{
					"type": "github",
					"value": map[string]any{
						"repo_address": dep.RepoPattern.Repo(),
						"github_repo":  dep.RepoPattern.GithubRepo(),
						"release":      dep.RepoPattern.Release(),
						"packages":     dep.RepoPattern.Asset(),
					},
				})
			} else if dep, ok := dependency.Value.(bundle_entities.MarketplaceDependency); ok {
				result = append(result, map[string]any{
					"type": "marketplace",
					"value": map[string]any{
						"organization": dep.MarketplacePattern.Organization(),
						"plugin":       dep.MarketplacePattern.Plugin(),
						"version":      dep.MarketplacePattern.Version(),
					},
				})
			} else if dep, ok := dependency.Value.(bundle_entities.PackageDependency); ok {
				// fetch package
				path := dep.Path
				if asset, err := packager.FetchAsset(path); err != nil {
					return entities.NewErrorResponse(-500, errors.Join(err, errors.New("failed to fetch package")).Error())
				} else {
					// decode and save
					decoder, err := decoder.NewZipPluginDecoder(asset)
					if err != nil {
						return entities.NewErrorResponse(-500, err.Error())
					}

					pluginUniqueIdentifier, err := decoder.UniqueIdentity()
					if err != nil {
						return entities.NewErrorResponse(-500, err.Error())
					}

					declaration, err := manager.SavePackage(pluginUniqueIdentifier, asset)
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

					result = append(result, map[string]any{
						"type": "package",
						"value": map[string]any{
							"unique_identifier": pluginUniqueIdentifier,
							"manifest":          declaration,
						},
					})
				}
			}
		}
	}

	return entities.NewSuccessResponse(result)
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

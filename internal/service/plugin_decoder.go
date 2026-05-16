package service

import (
	"errors"
	"io"
	"mime/multipart"

	"github.com/gin-gonic/gin"
	"github.com/langgenius/dify-plugin-daemon/internal/core/plugin_manager"
	"github.com/langgenius/dify-plugin-daemon/internal/types/app"
	"github.com/langgenius/dify-plugin-daemon/internal/types/exception"
	"github.com/langgenius/dify-plugin-daemon/pkg/bundle_packager"
	"github.com/langgenius/dify-plugin-daemon/pkg/entities"
	"github.com/langgenius/dify-plugin-daemon/pkg/entities/bundle_entities"
	"github.com/langgenius/dify-plugin-daemon/pkg/entities/plugin_entities"
	"github.com/langgenius/dify-plugin-daemon/pkg/plugin_packager/decoder"
	"github.com/langgenius/dify-plugin-daemon/pkg/utils/cache/helper"
)

func UploadPluginPkg(
	config *app.Config,
	c *gin.Context,
	tenantId string,
	difyPkgFile multipart.File,
	verifySignature bool,
) *entities.Response {
	pluginFile, err := io.ReadAll(difyPkgFile)
	if err != nil {
		return exception.InternalServerError(err).ToResponse()
	}

	decoderInstance, err := decoder.NewZipPluginDecoderWithSizeLimit(pluginFile, config.MaxPluginPackageSize)
	if err != nil {
		return exception.BadRequestError(err).ToResponse()
	}
	defer decoderInstance.Close()

	pluginUniqueIdentifier, err := decoderInstance.UniqueIdentity()
	if err != nil {
		return exception.BadRequestError(err).ToResponse()
	}

	// avoid author to be a uuid
	if pluginUniqueIdentifier.RemoteLike() {
		return exception.BadRequestError(errors.New("cannot set author to an UUID")).ToResponse()
	}

	manager := plugin_manager.Manager()
	preparedPackage, err := manager.PreparePackage(pluginFile, &decoder.ThirdPartySignatureVerificationConfig{
		Enabled:        config.ThirdPartySignatureVerificationEnabled,
		PublicKeyPaths: config.ThirdPartySignatureVerificationPublicKeys,
	})
	if err != nil {
		return exception.BadRequestError(errors.Join(err, errors.New("failed to decode package"))).ToResponse()
	}

	if config.ForceVerifyingSignature || verifySignature {
		if !preparedPackage.Declaration.Verified {
			return exception.BadRequestError(errors.Join(err, errors.New(
				"plugin verification has been enabled, and the plugin you want to install has a bad signature",
			))).ToResponse()
		}
	}

	declaration, err := manager.PersistPackage(preparedPackage)
	if err != nil {
		return exception.BadRequestError(errors.Join(err, errors.New("failed to save package"))).ToResponse()
	}

	// if config.EnforceLanggeniusSignatures {
	// 	if isUnauthorizedLanggenius(declaration, preparedPackage.Verification) {
	// 		return exception.BadRequestError(ErrUnauthorizedLanggenius).ToResponse()
	// 	}
	// }

	return entities.NewSuccessResponse(map[string]any{
		"unique_identifier": pluginUniqueIdentifier,
		"manifest":          declaration,
		"verification":      preparedPackage.Verification,
	})
}

func UploadPluginBundle(
	config *app.Config,
	c *gin.Context,
	tenant_id string,
	dify_bundle_file multipart.File,
	verify_signature bool,
) *entities.Response {
	bundleFile, err := io.ReadAll(dify_bundle_file)
	if err != nil {
		return exception.InternalServerError(err).ToResponse()
	}

	packager, err := bundle_packager.NewMemoryZipBundlePackager(bundleFile)
	if err != nil {
		return exception.BadRequestError(errors.Join(err, errors.New("failed to decode bundle"))).ToResponse()
	}

	// load bundle
	bundle, err := packager.Manifest()
	if err != nil {
		return exception.BadRequestError(errors.Join(err, errors.New("failed to load bundle manifest"))).ToResponse()
	}

	manager := plugin_manager.Manager()

	result := []map[string]any{}
	preparedPackageDependencies := []struct {
		resultIndex int
		prepared    *plugin_manager.PreparedPackage
	}{}

	for _, dependency := range bundle.Dependencies {
		if dependency.Type == bundle_entities.DEPENDENCY_TYPE_GITHUB {
			if dep, ok := dependency.Value.(bundle_entities.GithubDependency); ok {
				result = append(result, map[string]any{
					"type": "github",
					"value": map[string]any{
						"repo_address": dep.RepoPattern.Repo(),
						"repo":         dep.RepoPattern.GithubRepo(),
						"release":      dep.RepoPattern.Release(),
						"packages":     dep.RepoPattern.Asset(),
					},
				})
			}
		} else if dependency.Type == bundle_entities.DEPENDENCY_TYPE_MARKETPLACE {
			if dep, ok := dependency.Value.(bundle_entities.MarketplaceDependency); ok {
				result = append(result, map[string]any{
					"type": "marketplace",
					"value": map[string]any{
						"organization": dep.MarketplacePattern.Organization(),
						"plugin":       dep.MarketplacePattern.Plugin(),
						"version":      dep.MarketplacePattern.Version(),
					},
				})
			}
		} else if dependency.Type == bundle_entities.DEPENDENCY_TYPE_PACKAGE {
			if dep, ok := dependency.Value.(bundle_entities.PackageDependency); ok {
				// fetch package
				path := dep.Path
				if asset, err := packager.FetchAsset(path); err != nil {
					return exception.InternalServerError(errors.Join(errors.New("failed to fetch package from bundle"), err)).ToResponse()
				} else {
					// decode and save
					decoderInstance, err := decoder.NewZipPluginDecoder(asset)
					if err != nil {
						return exception.BadRequestError(errors.Join(errors.New("failed to create package decoder"), err)).ToResponse()
					}

					pluginUniqueIdentifier, err := decoderInstance.UniqueIdentity()
					_ = decoderInstance.Close()
					if err != nil {
						return exception.BadRequestError(errors.Join(errors.New("failed to get package unique identifier"), err)).ToResponse()
					}

					preparedPackage, err := manager.PreparePackage(asset, &decoder.ThirdPartySignatureVerificationConfig{
						Enabled:        config.ThirdPartySignatureVerificationEnabled,
						PublicKeyPaths: config.ThirdPartySignatureVerificationPublicKeys,
					})
					if err != nil {
						return exception.InternalServerError(errors.Join(errors.New("failed to decode package"), err)).ToResponse()
					}

					if config.ForceVerifyingSignature || verify_signature {
						if !preparedPackage.Declaration.Verified {
							return exception.BadRequestError(errors.Join(errors.New(
								"plugin verification has been enabled, and the plugin you want to install has a bad signature",
							), err)).ToResponse()
						}
					}

					if config.EnforceLanggeniusSignatures {
						if isUnauthorizedLanggenius(&preparedPackage.Declaration, preparedPackage.Verification) {
							return exception.BadRequestError(ErrUnauthorizedLanggenius).ToResponse()
						}
					}

					resultIndex := len(result)
					result = append(result, map[string]any{
						"type": "package",
						"value": map[string]any{
							"unique_identifier": pluginUniqueIdentifier,
							"manifest":          &preparedPackage.Declaration,
						},
					})
					preparedPackageDependencies = append(preparedPackageDependencies, struct {
						resultIndex int
						prepared    *plugin_manager.PreparedPackage
					}{resultIndex: resultIndex, prepared: preparedPackage})
				}
			}
		}
	}

	persistedPackages := []*plugin_manager.PreparedPackage{}
	for _, dependency := range preparedPackageDependencies {
		declaration, err := manager.PersistPackage(dependency.prepared)
		if err != nil {
			for i := len(persistedPackages) - 1; i >= 0; i-- {
				manager.RollbackPackage(persistedPackages[i])
			}
			return exception.InternalServerError(errors.Join(errors.New("failed to save package"), err)).ToResponse()
		}

		persistedPackages = append(persistedPackages, dependency.prepared)
		result[dependency.resultIndex] = map[string]any{
			"type": "package",
			"value": map[string]any{
				"unique_identifier": dependency.prepared.UniqueIdentifier,
				"manifest":          declaration,
			},
		}
	}

	return entities.NewSuccessResponse(result)
}

func FetchPluginManifest(
	pluginUniqueIdentifier plugin_entities.PluginUniqueIdentifier,
) *entities.Response {
	runtimeType := plugin_entities.PLUGIN_RUNTIME_TYPE_LOCAL
	if pluginUniqueIdentifier.RemoteLike() {
		runtimeType = plugin_entities.PLUGIN_RUNTIME_TYPE_REMOTE
	}

	pluginManifestCache, err := helper.CombinedGetPluginDeclaration(
		pluginUniqueIdentifier, runtimeType,
	)
	if err == helper.ErrPluginNotFound {
		return exception.BadRequestError(errors.New("plugin not found")).ToResponse()
	}

	if err != nil {
		return exception.InternalServerError(err).ToResponse()
	}

	return entities.NewSuccessResponse(pluginManifestCache)
}

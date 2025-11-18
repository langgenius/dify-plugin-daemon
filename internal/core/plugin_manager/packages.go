package plugin_manager

import (
	"errors"
	"fmt"
	"os"

	"github.com/langgenius/dify-plugin-daemon/internal/db"
	"github.com/langgenius/dify-plugin-daemon/internal/types/models"
	"github.com/langgenius/dify-plugin-daemon/pkg/entities/plugin_entities"
	"github.com/langgenius/dify-plugin-daemon/pkg/plugin_packager/decoder"
	"github.com/langgenius/dify-plugin-daemon/pkg/utils/cache/helper"
)

func (p *PluginManager) SavePackage(
	plugin_unique_identifier plugin_entities.PluginUniqueIdentifier,
	pkg []byte,
	thirdPartySignatureVerificationConfig *decoder.ThirdPartySignatureVerificationConfig,
) (*plugin_entities.PluginDeclaration, error) {
	// try to decode the package
	packageDecoder, err := decoder.NewZipPluginDecoderWithThirdPartySignatureVerificationConfig(pkg, thirdPartySignatureVerificationConfig)
	if err != nil {
		return nil, err
	}

	// get the declaration
	declaration, err := packageDecoder.Manifest()
	if err != nil {
		return nil, err
	}

	if err := declaration.ManifestValidate(); err != nil {
		return nil, errors.Join(err, fmt.Errorf("illegal plugin manifest"))
	}

	// get the assets
	assets, err := packageDecoder.Assets()
	if err != nil {
		return nil, err
	}

	// remap the assets
	_, err = p.mediaBucket.RemapAssets(&declaration, assets)
	if err != nil {
		return nil, errors.Join(err, fmt.Errorf("failed to remap assets"))
	}

	uniqueIdentifier, err := packageDecoder.UniqueIdentity()
	if err != nil {
		return nil, err
	}

	// save to storage
	err = p.packageBucket.Save(plugin_unique_identifier.String(), pkg)
	if err != nil {
		return nil, err
	}

	// create plugin if not exists
	if _, err := db.GetOne[models.PluginDeclaration](
		db.Equal("plugin_unique_identifier", uniqueIdentifier.String()),
	); err == db.ErrDatabaseNotFound {
		err = db.Create(&models.PluginDeclaration{
			PluginUniqueIdentifier: uniqueIdentifier.String(),
			PluginID:               uniqueIdentifier.PluginID(),
			Declaration:            declaration,
		})
		if err != nil {
			return nil, err
		}
	} else if err != nil {
		return nil, err
	}

	return &declaration, nil
}

func (p *PluginManager) GetPackage(
	plugin_unique_identifier plugin_entities.PluginUniqueIdentifier,
) ([]byte, error) {
	file, err := p.packageBucket.Get(plugin_unique_identifier.String())
	if err != nil {
		if os.IsNotExist(err) {
			return nil, errors.New("plugin package not found, please upload it firstly")
		}
		return nil, err
	}

	return file, nil
}

func (p *PluginManager) GetDeclaration(
	plugin_unique_identifier plugin_entities.PluginUniqueIdentifier,
	tenant_id string,
	runtime_type plugin_entities.PluginRuntimeType,
) (
	*plugin_entities.PluginDeclaration, error,
) {
	return helper.CombinedGetPluginDeclaration(
		plugin_unique_identifier, runtime_type,
	)
}

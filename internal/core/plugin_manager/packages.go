package plugin_manager

import (
	"bytes"
	"errors"
	"fmt"
	"os"
	"strings"

	"github.com/langgenius/dify-plugin-daemon/internal/db"
	"github.com/langgenius/dify-plugin-daemon/internal/types/models"
	"github.com/langgenius/dify-plugin-daemon/pkg/entities/plugin_entities"
	"github.com/langgenius/dify-plugin-daemon/pkg/plugin_packager/decoder"
	"github.com/langgenius/dify-plugin-daemon/pkg/utils/cache/helper"
)

type PreparedPackage struct {
	UniqueIdentifier plugin_entities.PluginUniqueIdentifier
	Declaration      plugin_entities.PluginDeclaration
	Verification     *decoder.Verification
	Package          []byte
	Assets           map[string][]byte
	AssetIDs         []string

	packageSaved       bool
	packageExisted     bool
	previousPackage    []byte
	declarationCreated bool
}

func (p *PluginManager) PreparePackage(
	pkg []byte,
	thirdPartySignatureVerificationConfig *decoder.ThirdPartySignatureVerificationConfig,
) (*PreparedPackage, error) {
	// try to decode the package
	packageDecoder, err := decoder.NewZipPluginDecoderWithThirdPartySignatureVerificationConfig(pkg, thirdPartySignatureVerificationConfig)
	if err != nil {
		return nil, err
	}
	defer packageDecoder.Close()

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

	uniqueIdentifier, err := packageDecoder.UniqueIdentity()
	if err != nil {
		return nil, err
	}

	verification, _ := packageDecoder.Verification()
	if verification == nil && packageDecoder.Verified() {
		verification = decoder.DefaultVerification()
	}

	return &PreparedPackage{
		UniqueIdentifier: uniqueIdentifier,
		Declaration:      declaration,
		Verification:     verification,
		Package:          pkg,
		Assets:           assets,
	}, nil
}

func (p *PluginManager) PersistPackage(
	prepared *PreparedPackage,
) (*plugin_entities.PluginDeclaration, error) {
	declaration := prepared.Declaration
	prepared.AssetIDs = nil
	prepared.packageSaved = false
	prepared.packageExisted = false
	prepared.previousPackage = nil
	prepared.declarationCreated = false

	packageExists, err := p.packageBucket.Exists(prepared.UniqueIdentifier.String())
	if err != nil {
		return nil, err
	}
	if packageExists {
		existingPackage, err := p.packageBucket.Get(prepared.UniqueIdentifier.String())
		if err != nil {
			return nil, err
		}
		prepared.packageExisted = true
		prepared.previousPackage = existingPackage
	}

	// remap the assets
	assetIDs, err := p.mediaBucket.RemapAssets(&declaration, prepared.Assets)
	if err != nil {
		prepared.AssetIDs = assetIDs
		p.rollbackPreparedPackage(prepared)
		return nil, errors.Join(err, fmt.Errorf("failed to remap assets"))
	}
	prepared.AssetIDs = assetIDs

	// save to storage
	err = p.packageBucket.Save(prepared.UniqueIdentifier.String(), prepared.Package)
	if err != nil {
		p.rollbackPreparedPackage(prepared)
		return nil, err
	}
	prepared.packageSaved = true

	// create plugin if not exists (idempotent under concurrency)
	if _, err := db.GetOne[models.PluginDeclaration](
		db.Equal("plugin_unique_identifier", prepared.UniqueIdentifier.String()),
	); err == db.ErrDatabaseNotFound {
		createErr := db.Create(&models.PluginDeclaration{
			PluginUniqueIdentifier: prepared.UniqueIdentifier.String(),
			PluginID:               prepared.UniqueIdentifier.PluginID(),
			Declaration:            declaration,
		})
		if createErr != nil {
			// ignore Postgres unique-violation (23505) errors triggered by concurrent inserts
			if isUniqueViolation(createErr) {
				prepared.Declaration = declaration
				return &prepared.Declaration, nil
			}
			// fallback: if another goroutine has just inserted, read-after-write should succeed
			if _, again := db.GetOne[models.PluginDeclaration](
				db.Equal("plugin_unique_identifier", prepared.UniqueIdentifier.String()),
			); again == nil {
				prepared.Declaration = declaration
				return &prepared.Declaration, nil
			}
			p.rollbackPreparedPackage(prepared)
			return nil, createErr
		}
		prepared.declarationCreated = true
	} else if err != nil {
		p.rollbackPreparedPackage(prepared)
		return nil, err
	}

	prepared.Declaration = declaration
	return &prepared.Declaration, nil
}

func (p *PluginManager) SavePackage(
	plugin_unique_identifier plugin_entities.PluginUniqueIdentifier,
	pkg []byte,
	thirdPartySignatureVerificationConfig *decoder.ThirdPartySignatureVerificationConfig,
) (*plugin_entities.PluginDeclaration, error) {
	prepared, err := p.PreparePackage(pkg, thirdPartySignatureVerificationConfig)
	if err != nil {
		return nil, err
	}

	if prepared.UniqueIdentifier != plugin_unique_identifier {
		return nil, fmt.Errorf("plugin unique identifier mismatch")
	}

	return p.PersistPackage(prepared)
}

func (p *PluginManager) RollbackPackage(prepared *PreparedPackage) {
	p.rollbackPreparedPackage(prepared)
}

func (p *PluginManager) rollbackPreparedPackage(prepared *PreparedPackage) {
	for _, assetID := range prepared.AssetIDs {
		_ = p.mediaBucket.Delete(assetID)
	}
	if prepared.declarationCreated {
		_ = db.DeleteByCondition(models.PluginDeclaration{
			PluginUniqueIdentifier: prepared.UniqueIdentifier.String(),
		})
	}
	if prepared.packageSaved {
		currentPackage, err := p.packageBucket.Get(prepared.UniqueIdentifier.String())
		if err == nil && bytes.Equal(currentPackage, prepared.Package) {
			if _, dbErr := db.GetOne[models.PluginDeclaration](
				db.Equal("plugin_unique_identifier", prepared.UniqueIdentifier.String()),
			); dbErr == db.ErrDatabaseNotFound {
				if prepared.packageExisted {
					_ = p.packageBucket.Save(prepared.UniqueIdentifier.String(), prepared.previousPackage)
				} else {
					_ = p.packageBucket.Delete(prepared.UniqueIdentifier.String())
				}
			}
		}
	}
}

// isUniqueViolation returns true if err indicates a PostgreSQL unique constraint violation (SQLSTATE 23505).
// Works across common drivers by matching canonical substrings to avoid hard dependency on driver types.
func isUniqueViolation(err error) bool {
	if err == nil {
		return false
	}
	s := err.Error()
	return strings.Contains(s, "SQLSTATE 23505") || strings.Contains(s, "duplicate key value violates unique constraint")
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

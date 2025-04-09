package bundle_packager

import (
	"archive/zip"
	"bytes"
	"errors"
	"os"
	"path/filepath"

	"github.com/langgenius/dify-plugin-daemon/internal/utils/parser"
	"github.com/langgenius/dify-plugin-daemon/pkg/entities/bundle_entities"
	"github.com/langgenius/dify-plugin-daemon/pkg/entities/manifest_entities"
	"github.com/langgenius/dify-plugin-daemon/pkg/plugin_packager/decoder"
)

type GenericBundlePackager struct {
	bundle *bundle_entities.Bundle
	assets map[string]*bytes.Buffer

	extraFiles map[string]*bytes.Buffer
}

func NewGenericBundlePackager(
	bundle *bundle_entities.Bundle,
	extraFiles map[string]*bytes.Buffer,
) *GenericBundlePackager {
	return &GenericBundlePackager{
		bundle:     bundle,
		assets:     make(map[string]*bytes.Buffer),
		extraFiles: extraFiles,
	}
}

func (p *GenericBundlePackager) Export() ([]byte, error) {
	// build a new zip file
	buffer := bytes.NewBuffer([]byte{})
	zipWriter := zip.NewWriter(buffer)
	defer zipWriter.Close()

	// write the manifest file
	manifestFile, err := zipWriter.Create("manifest.yaml")
	if err != nil {
		return nil, err
	}

	manifestBytes := parser.MarshalYamlBytes(p.bundle)
	_, err = manifestFile.Write(manifestBytes)
	if err != nil {
		return nil, err
	}

	// write the assets
	for name, asset := range p.assets {
		assetFile, err := zipWriter.Create(filepath.Join("_assets", name))
		if err != nil {
			return nil, err
		}

		_, err = assetFile.Write(asset.Bytes())
		if err != nil {
			return nil, err
		}
	}

	// write the extra files
	for name, file := range p.extraFiles {
		extraFile, err := zipWriter.Create(name)
		if err != nil {
			return nil, err
		}

		_, err = extraFile.Write(file.Bytes())
		if err != nil {
			return nil, err
		}
	}

	// close the zip writer to flush the buffer
	zipWriter.Close()

	return buffer.Bytes(), nil
}

func (p *GenericBundlePackager) Manifest() (*bundle_entities.Bundle, error) {
	return p.bundle, nil
}

func (p *GenericBundlePackager) Regenerate(bundle bundle_entities.Bundle) error {
	// replace the basic information of the bundle
	p.bundle.Author = bundle.Author
	p.bundle.Description = bundle.Description
	p.bundle.Name = bundle.Name
	p.bundle.Labels = bundle.Labels

	return nil
}

func (p *GenericBundlePackager) AppendGithubDependency(repoPattern bundle_entities.GithubRepoPattern) {
	p.bundle.Dependencies = append(p.bundle.Dependencies, bundle_entities.Dependency{
		Type: bundle_entities.DEPENDENCY_TYPE_GITHUB,
		Value: bundle_entities.GithubDependency{
			RepoPattern: repoPattern,
		},
	})
}

func (p *GenericBundlePackager) AppendMarketplaceDependency(marketplacePattern bundle_entities.MarketplacePattern) {
	p.bundle.Dependencies = append(p.bundle.Dependencies, bundle_entities.Dependency{
		Type: bundle_entities.DEPENDENCY_TYPE_MARKETPLACE,
		Value: bundle_entities.MarketplaceDependency{
			MarketplacePattern: marketplacePattern,
		},
	})
}

func (p *GenericBundlePackager) AppendPackageDependency(packagePath string) error {
	// try to read the packagePath as a file
	file, err := os.ReadFile(packagePath)
	if err != nil {
		return err
	}

	// try decode the file as a zip file
	zipDecoder, err := decoder.NewZipPluginDecoder(file)
	if err != nil {
		return errors.Join(err, errors.New("please provider a valid difypkg file"))
	}

	checksum, err := zipDecoder.Checksum()
	if err != nil {
		return errors.Join(err, errors.New("failed to get checksum of the package"))
	}

	p.assets[checksum] = bytes.NewBuffer(file)
	p.bundle.Dependencies = append(p.bundle.Dependencies, bundle_entities.Dependency{
		Type: bundle_entities.DEPENDENCY_TYPE_PACKAGE,
		Value: bundle_entities.PackageDependency{
			Path: checksum,
		},
	})

	return nil
}

func (p *GenericBundlePackager) ListDependencies() ([]bundle_entities.Dependency, error) {
	return p.bundle.Dependencies, nil
}

func (p *GenericBundlePackager) Remove(index int) error {
	if index < 0 || index >= len(p.bundle.Dependencies) {
		return errors.New("index out of bounds")
	}

	// get the dependency
	dependency := p.bundle.Dependencies[index]

	// remove the asset
	p.bundle.Dependencies = append(p.bundle.Dependencies[:index], p.bundle.Dependencies[index+1:]...)

	// delete the asset
	depValue, ok := dependency.Value.(bundle_entities.PackageDependency)
	if ok {
		delete(p.assets, depValue.Path)
	}

	return nil
}

func (p *GenericBundlePackager) BumpVersion(target manifest_entities.Version) {
	p.bundle.Version = target
}

func (p *GenericBundlePackager) FetchAsset(path string) ([]byte, error) {
	asset, ok := p.assets[path]
	if !ok {
		return nil, errors.New("asset not found")
	}

	return asset.Bytes(), nil
}

func (p *GenericBundlePackager) Assets() (map[string][]byte, error) {
	assets := make(map[string][]byte)
	for path, asset := range p.assets {
		assets[path] = asset.Bytes()
	}
	return assets, nil
}

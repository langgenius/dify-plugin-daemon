package bundle_packager

import (
	"github.com/langgenius/dify-plugin-daemon/internal/types/entities/bundle_entities"
	"github.com/langgenius/dify-plugin-daemon/internal/types/entities/manifest_entities"
)

type BundlePackagerImpl struct {
	bundle *bundle_entities.Bundle
}

func NewBundlePackager(zip []byte) BundlePackager {
	return nil
}

func (p *BundlePackagerImpl) Export() ([]byte, error) {
	return nil, nil
}

func (p *BundlePackagerImpl) Icon() ([]byte, error) {
	return nil, nil
}

func (p *BundlePackagerImpl) Manifest() (*bundle_entities.Bundle, error) {
	return p.bundle, nil
}

func (p *BundlePackagerImpl) Regenerate(bundle bundle_entities.Bundle) error {
	return nil
}

func (p *BundlePackagerImpl) AppendGithubDependency(repoPattern bundle_entities.GithubRepoPattern) {

}

func (p *BundlePackagerImpl) AppendMarketplaceDependency(marketplacePattern bundle_entities.MarketplacePattern) {

}

func (p *BundlePackagerImpl) AppendPackageDependency(packagePath string) {

}

func (p *BundlePackagerImpl) ListDependencies() ([]bundle_entities.Dependency, error) {
	return nil, nil
}

func (p *BundlePackagerImpl) Remove(index int) error {
	return nil
}

func (p *BundlePackagerImpl) BumpVersion(target manifest_entities.Version) {

}

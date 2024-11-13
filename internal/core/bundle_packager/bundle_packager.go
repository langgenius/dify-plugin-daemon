package bundle_packager

import (
	"github.com/langgenius/dify-plugin-daemon/internal/types/entities/bundle_entities"
)

type BundlePackager interface {
	// Export exports the bundle to a zip byte array
	Export() ([]byte, error)

	// Icon returns the icon of the bundle
	Icon() ([]byte, error)

	// Manifest returns the manifest of the bundle
	Manifest() (*bundle_entities.Bundle, error)

	// Remove removes a dependency from the bundle
	Remove(index int) error

	// Append Github Dependency appends a github dependency to the bundle
	AppendGithubDependency(repoPattern bundle_entities.GithubRepoPattern)

	// Append Marketplace Dependency appends a marketplace dependency to the bundle
	AppendMarketplaceDependency(marketplacePattern bundle_entities.MarketplacePattern)

	// Append Package Dependency appends a local package dependency to the bundle
	AppendPackageDependency(packagePath string)

	// ListDependencies lists all the dependencies of the bundle
	ListDependencies() ([]bundle_entities.Dependency, error)

	// Regenerate regenerates the bundle, replace the basic information of the bundle like name, labels, description, icon, etc.
	Regenerate(bundle bundle_entities.Bundle) error
}

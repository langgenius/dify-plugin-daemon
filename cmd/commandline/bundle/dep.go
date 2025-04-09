package bundle

import (
	"os"

	"github.com/langgenius/dify-plugin-daemon/internal/utils/log"
	"github.com/langgenius/dify-plugin-daemon/pkg/bundle_packager"
	"github.com/langgenius/dify-plugin-daemon/pkg/entities/bundle_entities"
)

func loadBundlePackager(bundlePath string) (bundle_packager.BundlePackager, error) {
	// state file, check if it's a file or a directory
	stateFile, err := os.Stat(bundlePath)
	if err != nil {
		return nil, err
	}

	if stateFile.IsDir() {
		return bundle_packager.NewLocalBundlePackager(bundlePath)
	}

	return bundle_packager.NewZipBundlePackager(bundlePath)
}

func AddGithubDependency(bundlePath string, pattern bundle_entities.GithubRepoPattern) {
	packager, err := loadBundlePackager(bundlePath)
	if err != nil {
		log.Error("Failed to load bundle packager: %v", err)
		return
	}

	packager.AppendGithubDependency(pattern)
	if err := packager.Save(); err != nil {
		log.Error("Failed to save bundle packager: %v", err)
		return
	}

	log.Info("Successfully added github dependency")
}

func AddMarketplaceDependency(bundlePath string, pattern bundle_entities.MarketplacePattern) {
	packager, err := loadBundlePackager(bundlePath)
	if err != nil {
		log.Error("Failed to load bundle packager: %v", err)
		return
	}

	packager.AppendMarketplaceDependency(pattern)
	if err := packager.Save(); err != nil {
		log.Error("Failed to save bundle packager: %v", err)
		return
	}

	log.Info("Successfully added marketplace dependency")
}

func AddPackageDependency(bundlePath string, path string) {
	packager, err := loadBundlePackager(bundlePath)
	if err != nil {
		log.Error("Failed to load bundle packager: %v", err)
		return
	}

	if err := packager.AppendPackageDependency(path); err != nil {
		log.Error("Failed to append package dependency: %v", err)
		return
	}

	if err := packager.Save(); err != nil {
		log.Error("Failed to save bundle packager: %v", err)
		return
	}

	log.Info("Successfully added package dependency")
}

func RegenerateBundle(bundlePath string) {
	bundle, err := generateNewBundle()
	if err != nil {
		log.Error("Failed to generate new bundle: %v", err)
		return
	}

	packager, err := loadBundlePackager(bundlePath)
	if err != nil {
		log.Error("Failed to load bundle packager: %v", err)
		return
	}

	packager.Regenerate(*bundle)
	if err := packager.Save(); err != nil {
		log.Error("Failed to save bundle packager: %v", err)
		return
	}

	log.Info("Successfully regenerated bundle")
}

func RemoveDependency(bundlePath string, index int) {
	packager, err := loadBundlePackager(bundlePath)
	if err != nil {
		log.Error("Failed to load bundle packager: %v", err)
		return
	}

	if err := packager.Remove(index); err != nil {
		log.Error("Failed to remove dependency: %v", err)
		return
	}

	if err := packager.Save(); err != nil {
		log.Error("Failed to save bundle packager: %v", err)
		return
	}

	log.Info("Successfully removed dependency")
}

func ListDependencies(bundlePath string) {
	packager, err := loadBundlePackager(bundlePath)
	if err != nil {
		log.Error("Failed to load bundle packager: %v", err)
		return
	}

	dependencies, err := packager.ListDependencies()
	if err != nil {
		log.Error("Failed to list dependencies: %v", err)
		return
	}

	if len(dependencies) == 0 {
		log.Info("No dependencies found")
		return
	}

	for i, dependency := range dependencies {
		log.Info("========== Dependency %d ==========", i)
		if dependency.Type == bundle_entities.DEPENDENCY_TYPE_GITHUB {
			githubDependency, ok := dependency.Value.(bundle_entities.GithubDependency)
			if !ok {
				log.Error("Failed to assert github pattern")
				continue
			}

			log.Info("Dependency Type: Github, Pattern: %s", githubDependency.RepoPattern)
			log.Info("Github Repo: %s", githubDependency.RepoPattern.Repo())
			log.Info("Release: %s", githubDependency.RepoPattern.Release())
			log.Info("Asset: %s", githubDependency.RepoPattern.Asset())
		} else if dependency.Type == bundle_entities.DEPENDENCY_TYPE_MARKETPLACE {
			marketplaceDependency, ok := dependency.Value.(bundle_entities.MarketplaceDependency)
			if !ok {
				log.Error("Failed to assert marketplace pattern")
				continue
			}

			log.Info("Dependency Type: Marketplace, Pattern: %s", marketplaceDependency.MarketplacePattern)
			log.Info("Organization: %s", marketplaceDependency.MarketplacePattern.Organization())
			log.Info("Plugin: %s", marketplaceDependency.MarketplacePattern.Plugin())
			log.Info("Version: %s", marketplaceDependency.MarketplacePattern.Version())
		} else if dependency.Type == bundle_entities.DEPENDENCY_TYPE_PACKAGE {
			packageDependency, ok := dependency.Value.(bundle_entities.PackageDependency)
			if !ok {
				log.Error("Failed to assert package dependency")
				continue
			}

			log.Info("Dependency Type: Package, Path: %s", packageDependency.Path)
			if asset, err := packager.FetchAsset(packageDependency.Path); err != nil {
				log.Error("Package %s not found", packageDependency.Path)
			} else {
				log.Info("Package %s: %d bytes", packageDependency.Path, len(asset))
			}
		}
	}
}

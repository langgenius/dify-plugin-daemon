package bundle

import (
	"os"

	"github.com/langgenius/dify-plugin-daemon/pkg/bundle_packager"
	"github.com/langgenius/dify-plugin-daemon/pkg/entities/bundle_entities"
	"github.com/langgenius/dify-plugin-daemon/pkg/utils/log"
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
		log.Error("failed to load bundle packager", "error", err)
		return
	}

	packager.AppendGithubDependency(pattern)
	if err := packager.Save(); err != nil {
		log.Error("failed to save bundle packager", "error", err)
		return
	}

	log.Info("successfully added github dependency")
}

func AddMarketplaceDependency(bundlePath string, pattern bundle_entities.MarketplacePattern) {
	packager, err := loadBundlePackager(bundlePath)
	if err != nil {
		log.Error("failed to load bundle packager", "error", err)
		return
	}

	packager.AppendMarketplaceDependency(pattern)
	if err := packager.Save(); err != nil {
		log.Error("failed to save bundle packager", "error", err)
		return
	}

	log.Info("successfully added marketplace dependency")
}

func AddPackageDependency(bundlePath string, path string) {
	packager, err := loadBundlePackager(bundlePath)
	if err != nil {
		log.Error("failed to load bundle packager", "error", err)
		return
	}

	if err := packager.AppendPackageDependency(path); err != nil {
		log.Error("failed to append package dependency", "error", err)
		return
	}

	if err := packager.Save(); err != nil {
		log.Error("failed to save bundle packager", "error", err)
		return
	}

	log.Info("successfully added package dependency")
}

func RegenerateBundle(bundlePath string) {
	bundle, err := generateNewBundle()
	if err != nil {
		log.Error("failed to generate new bundle", "error", err)
		return
	}

	packager, err := loadBundlePackager(bundlePath)
	if err != nil {
		log.Error("failed to load bundle packager", "error", err)
		return
	}

	packager.Regenerate(*bundle)
	if err := packager.Save(); err != nil {
		log.Error("failed to save bundle packager", "error", err)
		return
	}

	log.Info("successfully regenerated bundle")
}

func RemoveDependency(bundlePath string, index int) {
	packager, err := loadBundlePackager(bundlePath)
	if err != nil {
		log.Error("failed to load bundle packager", "error", err)
		return
	}

	if err := packager.Remove(index); err != nil {
		log.Error("failed to remove dependency", "error", err)
		return
	}

	if err := packager.Save(); err != nil {
		log.Error("failed to save bundle packager", "error", err)
		return
	}

	log.Info("successfully removed dependency")
}

func ListDependencies(bundlePath string) {
	packager, err := loadBundlePackager(bundlePath)
	if err != nil {
		log.Error("failed to load bundle packager", "error", err)
		return
	}

	dependencies, err := packager.ListDependencies()
	if err != nil {
		log.Error("failed to list dependencies", "error", err)
		return
	}

	if len(dependencies) == 0 {
		log.Info("no dependencies found")
		return
	}

	for i, dependency := range dependencies {
		log.Info("dependency", "index", i)
		if dependency.Type == bundle_entities.DEPENDENCY_TYPE_GITHUB {
			githubDependency, ok := dependency.Value.(bundle_entities.GithubDependency)
			if !ok {
				log.Error("failed to assert github pattern")
				continue
			}

			log.Info("github dependency",
				"pattern", githubDependency.RepoPattern,
				"repo", githubDependency.RepoPattern.Repo(),
				"release", githubDependency.RepoPattern.Release(),
				"asset", githubDependency.RepoPattern.Asset())
		} else if dependency.Type == bundle_entities.DEPENDENCY_TYPE_MARKETPLACE {
			marketplaceDependency, ok := dependency.Value.(bundle_entities.MarketplaceDependency)
			if !ok {
				log.Error("failed to assert marketplace pattern")
				continue
			}

			log.Info("marketplace dependency",
				"pattern", marketplaceDependency.MarketplacePattern,
				"organization", marketplaceDependency.MarketplacePattern.Organization(),
				"plugin", marketplaceDependency.MarketplacePattern.Plugin(),
				"version", marketplaceDependency.MarketplacePattern.Version())
		} else if dependency.Type == bundle_entities.DEPENDENCY_TYPE_PACKAGE {
			packageDependency, ok := dependency.Value.(bundle_entities.PackageDependency)
			if !ok {
				log.Error("failed to assert package dependency")
				continue
			}

			if asset, err := packager.FetchAsset(packageDependency.Path); err != nil {
				log.Error("package not found", "path", packageDependency.Path)
			} else {
				log.Info("package dependency", "path", packageDependency.Path, "size", len(asset))
			}
		}
	}
}

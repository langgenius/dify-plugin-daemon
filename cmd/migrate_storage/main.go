package main

import (
	"flag"
	"fmt"
	"os"
	"path"
	"strings"
	"time"

	"github.com/joho/godotenv"
	"github.com/kelseyhightower/envconfig"
	"github.com/langgenius/dify-cloud-kit/oss"
    "github.com/langgenius/dify-plugin-daemon/internal/storagefactory"
	"github.com/langgenius/dify-plugin-daemon/internal/types/app"
	"github.com/langgenius/dify-plugin-daemon/internal/utils/log"
)

// migrateCategory represents a named subpath that we copy
type migrateCategory struct {
	name string
	path string
}

// copyPrefix recursively copies files under a given prefix from src to dst.
func copyPrefix(src, dst oss.OSS, prefix string, dryRun bool) (files, skipped int, err error) {
	// simple BFS traversal using a queue of prefixes
	queue := []string{prefix}

	for len(queue) > 0 {
		current := queue[0]
		queue = queue[1:]

		// List current prefix
		entries, listErr := src.List(current)
		if listErr != nil {
			return files, skipped, fmt.Errorf("list %s failed: %w", current, listErr)
		}

		for _, e := range entries {
			// e.Path is the full path relative to the storage root
			if e.IsDir {
				next := e.Path
				if !strings.HasPrefix(next, current+"/") && next != current {
					next = path.Join(current, next)
				}
				queue = append(queue, next)
				continue
			}

			// skip dot files
			base := e.Path
			if strings.HasPrefix(base, ".") || strings.Contains(base, "/.") {
				skipped++
				continue
			}

			// check if exists at destination
			key := e.Path
			if !strings.HasPrefix(key, current+"/") && key != current {
				key = path.Join(current, key)
			}
			exists, exErr := dst.Exists(key)
			if exErr == nil && exists {
				skipped++
				continue
			}

			if dryRun {
				log.Info("DRYRUN copy %s", key)
				files++
				continue
			}

			// load and save
			data, loadErr := src.Load(key)
			if loadErr != nil {
				return files, skipped, fmt.Errorf("load %s failed: %w", key, loadErr)
			}
			if saveErr := dst.Save(key, data); saveErr != nil {
				return files, skipped, fmt.Errorf("save %s failed: %w", key, saveErr)
			}
			files++
		}
	}

	return files, skipped, nil
}

func main() {
	// Load .env if present
	_ = godotenv.Load()

	// CLI flags
	var (
		sourceRootOverride string
		only               string
		dryRun             bool
	)
	flag.StringVar(&sourceRootOverride, "source-root", "", "override PLUGIN_STORAGE_LOCAL_ROOT (default reads from .env)")
	flag.StringVar(&only, "only", "", "comma-separated categories to migrate: packages,assets,installed")
	flag.BoolVar(&dryRun, "dry-run", false, "list actions without uploading")
	flag.Parse()

	// Read config from env
	var cfg app.Config
	if err := envconfig.Process("", &cfg); err != nil {
		log.Panic("Error processing environment: %s", err.Error())
	}
	cfg.SetDefault()

	// We don't need full Validate here; allow PLATFORM local/serverless etc.
	// But ensure required pieces exist for destination
	if cfg.PluginStorageType == "" {
		log.Panic("DEST PLUGIN_STORAGE_TYPE is empty in env")
	}
	// Restrict: source must be local and destination must be cloud (non-local)
	if cfg.PluginStorageType == oss.OSS_TYPE_LOCAL {
		log.Panic("Destination PLUGIN_STORAGE_TYPE must be non-local (cloud). Local→Local migration is not allowed")
	}

	// Override local root if provided
	if sourceRootOverride != "" {
		cfg.PluginStorageLocalRoot = sourceRootOverride
	}
	if cfg.PluginStorageLocalRoot == "" {
		cfg.PluginStorageLocalRoot = "storage"
	}

	// Build source (local) and destination (cloud) storage
    // Storage construction is centralized in internal/storagefactory.New
    src, err := storagefactory.New(oss.OSS_TYPE_LOCAL, &cfg)
	if err != nil {
		log.Panic("Init source(local) storage failed: %s", err.Error())
	}
    dst, err := storagefactory.New(cfg.PluginStorageType, &cfg)
	if err != nil {
		log.Panic("Init destination(%s) storage failed: %s", cfg.PluginStorageType, err.Error())
	}

	// categories
	cats := []migrateCategory{
		{name: "packages", path: cfg.PluginPackageCachePath},
		{name: "assets", path: cfg.PluginMediaCachePath},
		{name: "installed", path: cfg.PluginInstalledPath},
	}

	// filter by --only if provided
	if only != "" {
		allow := map[string]bool{}
		for _, p := range strings.Split(only, ",") {
			p = strings.TrimSpace(p)
			if p != "" {
				allow[p] = true
			}
		}
		filtered := make([]migrateCategory, 0, len(cats))
		for _, c := range cats {
			if allow[c.name] {
				filtered = append(filtered, c)
			}
		}
		cats = filtered
	}

	if len(cats) == 0 {
		fmt.Fprintln(os.Stderr, "nothing to migrate; check --only")
		os.Exit(1)
	}

	start := time.Now()
	log.Info("Starting migration from local '%s' to '%s' bucket '%s'...", cfg.PluginStorageLocalRoot, cfg.PluginStorageType, cfg.PluginStorageOSSBucket)

	totalFiles := 0
	totalSkipped := 0
	for _, c := range cats {
		log.Info("Migrating %s (%s)...", c.name, c.path)
		n, s, err := copyPrefix(src, dst, c.path, dryRun)
		if err != nil {
			log.Panic("migrate %s failed: %s", c.name, err.Error())
		}
		totalFiles += n
		totalSkipped += s
		log.Info("Done %s: copied=%d skipped=%d", c.name, n, s)
	}

	dur := time.Since(start)
	log.Info("Migration completed in %s. Copied=%d Skipped=%d", dur.String(), totalFiles, totalSkipped)
}

package plugin

import (
	"bufio"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/joho/godotenv"
	"github.com/kelseyhightower/envconfig"
	"github.com/langgenius/dify-plugin-daemon/internal/db"
	"github.com/langgenius/dify-plugin-daemon/internal/types/app"
	"github.com/langgenius/dify-plugin-daemon/internal/types/models"
	"github.com/langgenius/dify-plugin-daemon/pkg/entities/manifest_entities"
	"github.com/langgenius/dify-plugin-daemon/pkg/plugin_packager/decoder"
	"github.com/langgenius/dify-plugin-daemon/pkg/utils/log"
)

type pluginDirEntry struct {
	PluginID string // author/name
	Version  string // x.y.z(-suffix)
	Checksum string
	DirName  string // on-disk folder name (identity-with-dashes@checksum)
	FullPath string
}

type versionTriple struct {
	major int
	minor int
	patch int
	// suffix like -alpha; empty means stable
	suffix string
}

func parseVersion(v string) (versionTriple, error) {
	// v should satisfy manifest_entities.VERSION_PATTERN: x.x.x or x.x.x-xxx
	// We will ignore any extra dots beyond 3 because regex guarantees exactly 3.
	main := v
	suffix := ""
	if i := strings.IndexRune(v, '-'); i >= 0 {
		main = v[:i]
		suffix = v[i+1:]
	}
	parts := strings.Split(main, ".")
	if len(parts) != 3 {
		return versionTriple{}, fmt.Errorf("invalid version: %s", v)
	}
	var out versionTriple
	if _, err := fmt.Sscanf(parts[0], "%d", &out.major); err != nil {
		return versionTriple{}, err
	}
	if _, err := fmt.Sscanf(parts[1], "%d", &out.minor); err != nil {
		return versionTriple{}, err
	}
	if _, err := fmt.Sscanf(parts[2], "%d", &out.patch); err != nil {
		return versionTriple{}, err
	}
	out.suffix = suffix
	return out, nil
}

func compareVersion(a, b string) int {
	// returns -1 if a<b, 0 if equal, 1 if a>b
	// Validate with regex; if invalid, treat as lowest.
	if !manifest_entities.PluginDeclarationVersionRegex.MatchString(a) && !manifest_entities.PluginDeclarationVersionRegex.MatchString(b) {
		return 0
	}
	if !manifest_entities.PluginDeclarationVersionRegex.MatchString(a) {
		return -1
	}
	if !manifest_entities.PluginDeclarationVersionRegex.MatchString(b) {
		return 1
	}
	va, _ := parseVersion(a)
	vb, _ := parseVersion(b)
	if va.major != vb.major {
		if va.major < vb.major {
			return -1
		}
		return 1
	}
	if va.minor != vb.minor {
		if va.minor < vb.minor {
			return -1
		}
		return 1
	}
	if va.patch != vb.patch {
		if va.patch < vb.patch {
			return -1
		}
		return 1
	}
	// Prefer no suffix over suffix when numbers equal
	sa := va.suffix
	sb := vb.suffix
	if sa == sb {
		return 0
	}
	if sa == "" {
		return 1
	}
	if sb == "" {
		return -1
	}
	// both have suffix: fall back to lexicographic
	if sa < sb {
		return -1
	}
	if sa > sb {
		return 1
	}
	return 0
}

// CleanupWorkingPath scans workingPath, groups plugin directories by pluginID, keeps latest version(s), and deletes older ones.
// If dryRun is true, only prints what would be deleted. If autoYes is false, ask for interactive confirmation before deletion.
// Also removes database rows that reference the deleted plugin unique identifiers.
func CleanupWorkingPath(workingPath string, dryRun bool, autoYes bool) {
	absWorkingPath, err := filepath.Abs(workingPath)
	if err != nil {
		log.Error("failed to get absolute working path", "error", err)
		return
	}

	// Collect plugin entries by reading only two levels: <working>/<author>/<name-version@checksum>
	byID := map[string][]pluginDirEntry{}

	authors, err := os.ReadDir(absWorkingPath)
	if err != nil {
		log.Error("failed to read working path", "path", absWorkingPath, "error", err)
		return
	}
	for _, a := range authors {
		if !a.IsDir() {
			continue
		}
		author := a.Name()
		authorDir := filepath.Join(absWorkingPath, author)
		entries, err := os.ReadDir(authorDir)
		if err != nil {
			log.Warn("failed to read author dir", "author", author, "error", err)
			continue
		}
		for _, e := range entries {
			if !e.IsDir() {
				continue
			}
			name := e.Name() // e.g. openai-0.2.7@abc...
			fullPath := filepath.Join(authorDir, name)
			pluginID, ver, checksum, ok := fastParseFromDir(author, name)
			if !ok {
				// fallback: verify plugin by manifest only if manifest.yaml exists
				manifestPath := filepath.Join(fullPath, "manifest.yaml")
				if _, statErr := os.Stat(manifestPath); statErr != nil {
					continue
				}
				dec, derr := decoder.NewFSPluginDecoder(fullPath)
				if derr != nil {
					log.Warn("skip invalid plugin directory", "path", fullPath, "error", derr)
					continue
				}
				manifest, derr := dec.Manifest()
				if derr != nil {
					log.Warn("skip directory without valid manifest", "path", fullPath, "error", derr)
					continue
				}
				pluginID = fmt.Sprintf("%s/%s", manifest.Author, manifest.Name)
				ver = manifest.Version.String()
				// if checksum not parsed, try from dir suffix
				if checksum == "" {
					if parts := strings.SplitN(name, "@", 2); len(parts) == 2 {
						checksum = parts[1]
					}
				}
			}
			byID[pluginID] = append(byID[pluginID], pluginDirEntry{
				PluginID: pluginID,
				Version:  ver,
				Checksum: checksum,
				DirName:  name,
				FullPath: fullPath,
			})
		}
	}

	// Decide what to delete
	toDelete := make([]pluginDirEntry, 0)
	toKeep := make(map[string][]pluginDirEntry)
	for id, list := range byID {
		if len(list) <= 1 {
			toKeep[id] = list
			continue
		}
		// Find max version
		sort.SliceStable(list, func(i, j int) bool { return compareVersion(list[i].Version, list[j].Version) > 0 })
		latestVersion := list[0].Version
		keep := make([]pluginDirEntry, 0)
		for _, item := range list {
			if compareVersion(item.Version, latestVersion) == 0 {
				keep = append(keep, item)
			} else {
				toDelete = append(toDelete, item)
			}
		}
		toKeep[id] = keep
	}

	if len(byID) == 0 {
		fmt.Println("No plugins found under:", absWorkingPath)
		return
	}

	if dryRun {
		fmt.Printf("[dry-run] Will remove %d old version directories:\n", len(toDelete))
		for _, d := range toDelete {
			fmt.Printf("- %s %s (checksum %s) -> %s\n", d.PluginID, d.Version, d.Checksum, d.FullPath)
		}
		return
	}

	if len(toDelete) == 0 {
		fmt.Println("Nothing to delete. All plugins are already at their latest versions.")
		return
	}

	// Ask for confirmation unless autoYes
	if !autoYes {
		fmt.Printf("You are about to delete %d directories. Proceed? [y/N]: ", len(toDelete))
		reader := bufio.NewReader(os.Stdin)
		answer, _ := reader.ReadString('\n')
		answer = strings.ToLower(strings.TrimSpace(answer))
		if answer != "y" && answer != "yes" {
			fmt.Println("Aborted.")
			return
		}
	}

	// Perform deletion with safety check
	deletedUIDs := make([]string, 0, len(toDelete))
	for _, d := range toDelete {
		absDir, err := filepath.Abs(d.FullPath)
		if err != nil {
			log.Error("failed to resolve path", "path", d.FullPath, "error", err)
			continue
		}
		// Ensure deletion target is under working path
		if !strings.HasPrefix(absDir, absWorkingPath+string(os.PathSeparator)) && absDir != absWorkingPath {
			log.Error("skip deletion due to invalid path (outside working dir)", "path", absDir)
			continue
		}
		if err := os.RemoveAll(absDir); err != nil {
			log.Error("failed to remove directory", "path", absDir, "error", errors.Join(err, errors.New("failed to remove plugin directory")))
			continue
		}
		log.Info("removed old plugin version", "plugin", d.PluginID, "version", d.Version, "path", absDir)
		// build plugin unique identifier for DB cleanup if checksum present
		if d.Checksum != "" {
			deletedUIDs = append(deletedUIDs, fmt.Sprintf("%s:%s@%s", d.PluginID, d.Version, d.Checksum))
		}
	}

	// Delete database rows referencing these unique identifiers
	if len(deletedUIDs) > 0 {
		cleanupDatabase(deletedUIDs)
	}

	fmt.Println("Cleanup complete.")
}

// fastParseFromDir attempts to parse author/name, version and checksum from directory layout
// <author>/<name>-<version>@<checksum>.
// Returns ok=false when parsing fails.
func fastParseFromDir(author, dirName string) (pluginID, version, checksum string, ok bool) {
	checksum = ""
	// optional checksum
	left := dirName
	if parts := strings.SplitN(dirName, "@", 2); len(parts) == 2 {
		left, checksum = parts[0], parts[1]
	}
	// split name-version by last '-'
	i := strings.LastIndex(left, "-")
	if i <= 0 || i >= len(left)-1 {
		return "", "", "", false
	}
	name := left[:i]
	version = left[i+1:]
	if !manifest_entities.PluginDeclarationVersionRegex.MatchString(version) {
		return "", "", "", false
	}
	pluginID = fmt.Sprintf("%s/%s", author, name)
	return pluginID, version, checksum, true
}

// cleanupDatabase removes rows across related tables by plugin_unique_identifier values.
func cleanupDatabase(uniqueIDs []string) {
	// Load env and init DB like server main does
	var cfg app.Config
	_ = godotenv.Load()
	if err := envconfig.Process("", &cfg); err != nil {
		log.Error("db env config load failed, skip db cleanup", "error", err)
		return
	}
	cfg.SetDefault()
	if err := cfg.Validate(); err != nil {
		log.Error("db config invalid, skip db cleanup", "error", err)
		return
	}
	// init DB
	db.Init(&cfg)
	defer db.Close()

	for _, uid := range uniqueIDs {
		// best-effort deletes; continue on error
		if err := db.DeleteByCondition(models.Plugin{PluginUniqueIdentifier: uid}); err != nil && !errors.Is(err, db.ErrDatabaseNotFound) {
			log.Warn("failed to delete Plugin by uid", "uid", uid, "error", err)
		}
		if err := db.DeleteByCondition(models.PluginDeclaration{PluginUniqueIdentifier: uid}); err != nil && !errors.Is(err, db.ErrDatabaseNotFound) {
			log.Warn("failed to delete PluginDeclaration by uid", "uid", uid, "error", err)
		}
		if err := db.DeleteByCondition(models.ServerlessRuntime{PluginUniqueIdentifier: uid}); err != nil && !errors.Is(err, db.ErrDatabaseNotFound) {
			log.Warn("failed to delete ServerlessRuntime by uid", "uid", uid, "error", err)
		}
		if err := db.DeleteByCondition(models.PluginInstallation{PluginUniqueIdentifier: uid}); err != nil && !errors.Is(err, db.ErrDatabaseNotFound) {
			log.Warn("failed to delete PluginInstallation by uid", "uid", uid, "error", err)
		}
		if err := db.DeleteByCondition(models.ToolInstallation{PluginUniqueIdentifier: uid}); err != nil && !errors.Is(err, db.ErrDatabaseNotFound) {
			log.Warn("failed to delete ToolInstallation by uid", "uid", uid, "error", err)
		}
		if err := db.DeleteByCondition(models.AIModelInstallation{PluginUniqueIdentifier: uid}); err != nil && !errors.Is(err, db.ErrDatabaseNotFound) {
			log.Warn("failed to delete AIModelInstallation by uid", "uid", uid, "error", err)
		}
		if err := db.DeleteByCondition(models.DatasourceInstallation{PluginUniqueIdentifier: uid}); err != nil && !errors.Is(err, db.ErrDatabaseNotFound) {
			log.Warn("failed to delete DatasourceInstallation by uid", "uid", uid, "error", err)
		}
		if err := db.DeleteByCondition(models.AgentStrategyInstallation{PluginUniqueIdentifier: uid}); err != nil && !errors.Is(err, db.ErrDatabaseNotFound) {
			log.Warn("failed to delete AgentStrategyInstallation by uid", "uid", uid, "error", err)
		}
		if err := db.DeleteByCondition(models.TriggerInstallation{PluginUniqueIdentifier: uid}); err != nil && !errors.Is(err, db.ErrDatabaseNotFound) {
			log.Warn("failed to delete TriggerInstallation by uid", "uid", uid, "error", err)
		}
		if err := db.DeleteByCondition(models.PluginReadmeRecord{PluginUniqueIdentifier: uid}); err != nil && !errors.Is(err, db.ErrDatabaseNotFound) {
			log.Warn("failed to delete PluginReadmeRecord by uid", "uid", uid, "error", err)
		}
		log.Info("deleted database records for plugin version", "uid", uid)
	}
}

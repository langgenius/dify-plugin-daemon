package bundle_entities

import (
	"fmt"
	"regexp"

	"github.com/go-playground/validator/v10"
	"github.com/langgenius/dify-plugin-daemon/internal/types/entities/manifest_entities"
	"github.com/langgenius/dify-plugin-daemon/internal/types/validators"
)

type DependencyType string

const (
	DEPENDENCY_TYPE_GITHUB      DependencyType = "github"
	DEPENDENCY_TYPE_MARKETPLACE DependencyType = "marketplace"
	DEPENDENCY_TYPE_PACKAGE     DependencyType = "package"
)

type Dependency struct {
	Type  DependencyType `json:"type" yaml:"type" validate:"required,oneof=github marketplace package"`
	Value any            `json:"value" yaml:"value" validate:"required"`
}

type GithubRepoPattern string

func NewGithubRepoPattern(pattern string) (GithubRepoPattern, error) {
	if !GITHUB_DEPENDENCY_PATTERN_REGEX_COMPILED.MatchString(pattern) {
		return "", fmt.Errorf("invalid github repo pattern")
	}
	return GithubRepoPattern(pattern), nil
}

type MarketplacePattern string

func NewMarketplacePattern(pattern string) (MarketplacePattern, error) {
	if !MARKETPLACE_PATTERN_REGEX_COMPILED.MatchString(pattern) {
		return "", fmt.Errorf("invalid marketplace pattern")
	}
	return MarketplacePattern(pattern), nil
}

var (
	GITHUB_VERSION_PATTERN = fmt.Sprintf(
		`([~^]?%s|%s(\.%s){2}|%s-%s)`,
		manifest_entities.VERSION_PATTERN,
		manifest_entities.VERSION_X_PATTERN,
		manifest_entities.VERSION_X_PATTERN,
		manifest_entities.VERSION_PATTERN,
		manifest_entities.VERSION_PATTERN,
	)
	GITHUB_DEPENDENCY_PATTERN = fmt.Sprintf(`^[a-z0-9_-]{1,64}/[a-z0-9_-]{1,128}/%s/[^/]+$`, GITHUB_VERSION_PATTERN)

	MARKETPLACE_VERSION_PATTERN = fmt.Sprintf(
		`([~^]?%s|%s(\.%s){2}|%s-%s)`,
		manifest_entities.VERSION_PATTERN,
		manifest_entities.VERSION_X_PATTERN,
		manifest_entities.VERSION_X_PATTERN,
		manifest_entities.VERSION_PATTERN,
		manifest_entities.VERSION_PATTERN,
	)
	MARKETPLACE_DEPENDENCY_PATTERN = fmt.Sprintf(`^[a-z0-9_-]{1,64}/[a-z0-9_-]{1,128}/%s$`, MARKETPLACE_VERSION_PATTERN)
)

var (
	GITHUB_DEPENDENCY_PATTERN_REGEX_COMPILED = regexp.MustCompile(GITHUB_DEPENDENCY_PATTERN)
	MARKETPLACE_PATTERN_REGEX_COMPILED       = regexp.MustCompile(MARKETPLACE_DEPENDENCY_PATTERN)
)

func validateGithubDependencyPattern(fl validator.FieldLevel) bool {
	return GITHUB_DEPENDENCY_PATTERN_REGEX_COMPILED.MatchString(fl.Field().String())
}

func validateMarketplacePattern(fl validator.FieldLevel) bool {
	return MARKETPLACE_PATTERN_REGEX_COMPILED.MatchString(fl.Field().String())
}

func init() {
	validators.GlobalEntitiesValidator.RegisterValidation("github_dependency_pattern", validateGithubDependencyPattern)
	validators.GlobalEntitiesValidator.RegisterValidation("marketplace_pattern", validateMarketplacePattern)
}

type GithubDependency struct {
	// RepoPattern is the pattern of the repo, as for its content, at least one of the following patterns:
	// 1. owner/repo/1.0.0/aaa.difypkg
	// 2. owner/repo/1.0.0/*.difypkg
	// 3. owner/repo/1.x.x/aaa.difypkg
	// 4. owner/repo/^1.0.0/aaa.difypkg
	// 5. owner/repo/~1.0.0/aaa.difypkg
	// 6. owner/repo/1.0.0-2.0.0/aaa.difypkg
	// 7. owner/repo/1.0.0-beta/aaa.difypkg
	RepoPattern GithubRepoPattern `json:"repo_pattern" yaml:"repo_pattern" validate:"required,github_dependency_pattern"`
}

type MarketplaceDependency struct {
	// MarketplacePattern is the pattern of the marketplace, as for its content, at least one of the following patterns:
	// 1. org/plugin/1.0.0
	// 2. org/plugin/1.x.x
	// 3. org/plugin/^1.0.0
	// 4. org/plugin/~1.0.0
	// 5. org/plugin/1.0.0-2.0.0
	// 6. org/plugin/1.0.0-beta
	MarketplacePattern MarketplacePattern `json:"marketplace_pattern" yaml:"marketplace_pattern" validate:"required,marketplace_pattern"`
}

type PackageDependency struct {
	// refers to the path of difypkg file in assets
	Path string `json:"path" yaml:"path" validate:"required"`
}

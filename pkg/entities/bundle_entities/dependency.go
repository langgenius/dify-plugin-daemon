package bundle_entities

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/go-playground/validator/v10"
	"github.com/langgenius/dify-plugin-daemon/pkg/entities/manifest_entities"
	"github.com/langgenius/dify-plugin-daemon/pkg/validators"
	"gopkg.in/yaml.v3"
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

func (d *Dependency) UnmarshalYAML(node *yaml.Node) error {
	// try convert Value to GithubDependency, MarketplaceDependency, PackageDependency
	type alias struct {
		Type  DependencyType `json:"type" yaml:"type" validate:"required,oneof=github marketplace package"`
		Value yaml.Node      `json:"value" yaml:"value" validate:"required"`
	}

	var a alias
	if err := node.Decode(&a); err != nil {
		return err
	}

	d.Type = a.Type

	// try convert Value to GithubDependency, MarketplaceDependency, PackageDependency
	switch d.Type {
	case DEPENDENCY_TYPE_GITHUB:
		var value GithubDependency
		if err := a.Value.Decode(&value); err != nil {
			return err
		}
		d.Value = value
	case DEPENDENCY_TYPE_MARKETPLACE:
		var value MarketplaceDependency
		if err := a.Value.Decode(&value); err != nil {
			return err
		}
		d.Value = value
	case DEPENDENCY_TYPE_PACKAGE:
		var value PackageDependency
		if err := a.Value.Decode(&value); err != nil {
			return err
		}
		d.Value = value
	}

	return nil
}

type GithubRepoPattern string

func (p GithubRepoPattern) Split() []string {
	split := strings.Split(string(p), ":")
	// split again by "/"
	splits := []string{}
	for _, s := range split {
		splits = append(splits, strings.Split(s, "/")...)
	}

	return splits
}

func (p GithubRepoPattern) Repo() string {
	split := p.Split()
	if len(split) < 3 {
		return ""
	}

	organization, repo := split[0], split[1]

	return fmt.Sprintf("https://github.com/%s/%s", organization, repo)
}

func (p GithubRepoPattern) GithubRepo() string {
	split := p.Split()
	if len(split) < 3 {
		return ""
	}

	return fmt.Sprintf("%s/%s", split[0], split[1])
}

func (p GithubRepoPattern) Release() string {
	split := p.Split()
	if len(split) < 3 {
		return ""
	}

	return split[2]
}

func (p GithubRepoPattern) Asset() string {
	split := p.Split()
	if len(split) < 4 {
		return ""
	}

	return split[3]
}

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

func (p MarketplacePattern) Split() []string {
	split := strings.Split(string(p), ":")
	// split again by "/"
	splits := []string{}
	for _, s := range split {
		splits = append(splits, strings.Split(s, "/")...)
	}

	return splits
}

func (p MarketplacePattern) Organization() string {
	split := p.Split()
	if len(split) < 1 {
		return ""
	}

	return split[0]
}

func (p MarketplacePattern) Plugin() string {
	split := p.Split()
	if len(split) < 2 {
		return ""
	}

	return split[1]
}

func (p MarketplacePattern) Version() string {
	split := p.Split()
	if len(split) < 3 {
		return ""
	}

	return split[2]
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
	GITHUB_DEPENDENCY_PATTERN = fmt.Sprintf(`^[a-z0-9_-]{1,64}/[a-z0-9_-]{1,128}:%s/[^/]+$`, GITHUB_VERSION_PATTERN)

	MARKETPLACE_VERSION_PATTERN = fmt.Sprintf(
		`([~^]?%s|%s(\.%s){2}|%s-%s)`,
		manifest_entities.VERSION_PATTERN,
		manifest_entities.VERSION_X_PATTERN,
		manifest_entities.VERSION_X_PATTERN,
		manifest_entities.VERSION_PATTERN,
		manifest_entities.VERSION_PATTERN,
	)
	MARKETPLACE_DEPENDENCY_PATTERN = fmt.Sprintf(`^[a-z0-9_-]{1,64}/[a-z0-9_-]{1,128}:%s$`, MARKETPLACE_VERSION_PATTERN)
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

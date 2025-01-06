package main

import (
	"fmt"
	"path/filepath"
	"strconv"

	"github.com/langgenius/dify-plugin-daemon/cmd/commandline/bundle"
	"github.com/langgenius/dify-plugin-daemon/internal/utils/log"
	"github.com/langgenius/dify-plugin-daemon/pkg/entities/bundle_entities"
	"github.com/spf13/cobra"
)

var (
	bundleCreateCommand = &cobra.Command{
		Use:   "init",
		Short: "Create a bundle",
		Long:  "Create a bundle",
		Run: func(c *cobra.Command, args []string) {
			bundle.InitBundle()
		},
	}

	bundleAnalyzeCommand = &cobra.Command{
		Use:   "analyze [bundle_path]",
		Short: "List all dependencies",
		Long:  "List all dependencies in the specified bundle",
		Args:  cobra.ExactArgs(1),
		Run: func(c *cobra.Command, args []string) {
			bundlePath := args[0]
			bundle.ListDependencies(bundlePath)
		},
	}

	bundleAppendDependencyCommand = &cobra.Command{
		Use:   "append",
		Short: "Append a dependency",
		Long:  "Append a dependency",
	}

	bundleAppendGithubDependencyCommand = &cobra.Command{
		Use:   "github [bundle_path]",
		Short: "Append a github dependency",
		Long:  "Append a github dependency to the specified bundle",
		Args:  cobra.ExactArgs(1),
		Run: func(c *cobra.Command, args []string) {
			bundlePath := args[0]
			repoPattern := c.Flag("repo_pattern").Value.String()
			githubPattern, err := bundle_entities.NewGithubRepoPattern(repoPattern)
			if err != nil {
				log.Error("Invalid github repo pattern: %v", err)
				return
			}
			bundle.AddGithubDependency(bundlePath, githubPattern)
		},
	}

	bundleAppendMarketplaceDependencyCommand = &cobra.Command{
		Use:   "marketplace [bundle_path]",
		Short: "Append a marketplace dependency",
		Long:  "Append a marketplace dependency to the specified bundle",
		Args:  cobra.ExactArgs(1),
		Run: func(c *cobra.Command, args []string) {
			bundlePath := args[0]
			marketplacePatternString := c.Flag("marketplace_pattern").Value.String()
			marketplacePattern, err := bundle_entities.NewMarketplacePattern(marketplacePatternString)
			if err != nil {
				log.Error("Invalid marketplace pattern: %v", err)
				return
			}
			bundle.AddMarketplaceDependency(bundlePath, marketplacePattern)
		},
	}

	bundleAppendPackageDependencyCommand = &cobra.Command{
		Use:   "package [bundle_path]",
		Short: "Append a local package dependency",
		Long:  "Append a local package dependency to the specified bundle",
		Args:  cobra.ExactArgs(1),
		Run: func(c *cobra.Command, args []string) {
			bundlePath := args[0]
			packagePath := c.Flag("package_path").Value.String()
			bundle.AddPackageDependency(bundlePath, packagePath)
		},
	}

	bundleRegenerateCommand = &cobra.Command{
		Use:   "regenerate [bundle_path]",
		Short: "Regenerate the bundle",
		Long:  "Regenerate the specified bundle",
		Args:  cobra.ExactArgs(1),
		Run: func(c *cobra.Command, args []string) {
			bundlePath := args[0]
			bundle.RegenerateBundle(bundlePath)
		},
	}

	bundleRemoveDependencyCommand = &cobra.Command{
		Use:   "remove [bundle_path]",
		Short: "Remove a dependency",
		Long:  "Remove a dependency from the specified bundle",
		Args:  cobra.ExactArgs(1),
		Run: func(c *cobra.Command, args []string) {
			bundlePath := args[0]
			index := c.Flag("index").Value.String()
			indexInt, err := strconv.Atoi(index)
			if err != nil {
				log.Error("Invalid index: %v", err)
				return
			}
			bundle.RemoveDependency(bundlePath, indexInt)
		},
	}

	bundleBumpVersionCommand = &cobra.Command{
		Use:   "bump [bundle_path]",
		Short: "Bump the version of the bundle",
		Long:  "Bump the version of the specified bundle",
		Args:  cobra.ExactArgs(1),
		Run: func(c *cobra.Command, args []string) {
			bundlePath := args[0]
			targetVersion := c.Flag("target_version").Value.String()
			bundle.BumpVersion(bundlePath, targetVersion)
		},
	}

	bundlePackageCommand = &cobra.Command{
		Use:   "package [bundle_path]",
		Short: "Package the bundle",
		Long:  "Package the specified bundle",
		Args:  cobra.ExactArgs(1),
		Run: func(c *cobra.Command, args []string) {
			bundlePath := args[0]
			// using filename of input_path as output_path if not specified
			outputPath := ""

			if c.Flag("output_path").Value.String() != "" {
				outputPath = c.Flag("output_path").Value.String()
			} else {
				base := filepath.Base(bundlePath)
				if base == "." || base == "/" {
					fmt.Println("Error: invalid input path, you should specify the path outside of bundle directory")
					return
				}
				outputPath = base + ".difybndl"
			}

			bundle.PackageBundle(bundlePath, outputPath)
		},
	}
)

func init() {
	bundleCommand.AddCommand(bundleCreateCommand)
	bundleCommand.AddCommand(bundleAppendDependencyCommand)
	bundleAppendDependencyCommand.AddCommand(bundleAppendGithubDependencyCommand)
	bundleAppendDependencyCommand.AddCommand(bundleAppendMarketplaceDependencyCommand)
	bundleAppendDependencyCommand.AddCommand(bundleAppendPackageDependencyCommand)
	bundleCommand.AddCommand(bundleRemoveDependencyCommand)
	bundleCommand.AddCommand(bundleRegenerateCommand)
	bundleCommand.AddCommand(bundleBumpVersionCommand)
	bundleCommand.AddCommand(bundlePackageCommand)
	bundleCommand.AddCommand(bundleAnalyzeCommand)

	bundleAppendGithubDependencyCommand.Flags().StringP("repo_pattern", "r", "", "github repo pattern")
	bundleAppendGithubDependencyCommand.MarkFlagRequired("repo_pattern")

	bundleAppendMarketplaceDependencyCommand.Flags().StringP("marketplace_pattern", "m", "", "marketplace pattern")
	bundleAppendMarketplaceDependencyCommand.MarkFlagRequired("marketplace_pattern")

	bundleAppendPackageDependencyCommand.Flags().StringP("package_path", "p", "", "path to the package")
	bundleAppendPackageDependencyCommand.MarkFlagRequired("package_path")

	bundleRemoveDependencyCommand.Flags().StringP("index", "i", "", "index of the dependency")
	bundleRemoveDependencyCommand.MarkFlagRequired("index")

	bundleBumpVersionCommand.Flags().StringP("target_version", "t", "", "target version")
	bundleBumpVersionCommand.MarkFlagRequired("target_version")

	bundlePackageCommand.Flags().StringP("output_path", "o", "", "output path")
}

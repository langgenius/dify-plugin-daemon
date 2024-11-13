package main

import (
	"github.com/spf13/cobra"
)

var (
	bundleCreateCommand = &cobra.Command{
		Use:   "create",
		Short: "Create a bundle",
		Long:  "Create a bundle",
		Run: func(c *cobra.Command, args []string) {
		},
	}

	bundleAnalyzeCommand = &cobra.Command{
		Use:   "analyze",
		Short: "List all dependencies",
		Long:  "List all dependencies",
		Run: func(c *cobra.Command, args []string) {
		},
	}

	bundleAppendDependencyCommand = &cobra.Command{
		Use:   "append",
		Short: "Append a dependency",
		Long:  "Append a dependency",
	}

	bundleAppendGithubDependencyCommand = &cobra.Command{
		Use:   "github",
		Short: "Append a github dependency",
		Long:  "Append a github dependency",
		Run: func(c *cobra.Command, args []string) {
		},
	}

	bundleAppendMarketplaceDependencyCommand = &cobra.Command{
		Use:   "marketplace",
		Short: "Append a marketplace dependency",
		Long:  "Append a marketplace dependency",
		Run: func(c *cobra.Command, args []string) {
		},
	}

	bundleAppendPackageDependencyCommand = &cobra.Command{
		Use:   "package",
		Short: "Append a local package dependency",
		Long:  "Append a local package dependency",
		Run: func(c *cobra.Command, args []string) {
		},
	}

	bundleRegenerateCommand = &cobra.Command{
		Use:   "regenerate",
		Short: "Regenerate the bundle",
		Long:  "Regenerate the bundle",
		Run: func(c *cobra.Command, args []string) {
		},
	}

	bundleRemoveDependencyCommand = &cobra.Command{
		Use:   "remove",
		Short: "Remove a dependency",
		Long:  "Remove a dependency",
		Run: func(c *cobra.Command, args []string) {
		},
	}

	bundleBumpVersionCommand = &cobra.Command{
		Use:   "bump",
		Short: "Bump the version of the bundle",
		Long:  "Bump the version of the bundle",
		Run: func(c *cobra.Command, args []string) {
		},
	}

	bundleListDependenciesCommand = &cobra.Command{
		Use:   "list",
		Short: "List all dependencies",
		Long:  "List all dependencies",
		Run: func(c *cobra.Command, args []string) {
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
	bundleCommand.AddCommand(bundleListDependenciesCommand)

	bundleCommand.AddCommand(bundleAnalyzeCommand)

	bundleAppendDependencyCommand.Flags().StringP("bundle_path", "i", "", "path to the bundle file")
	bundleAppendDependencyCommand.MarkFlagRequired("bundle_path")

	bundleAppendGithubDependencyCommand.Flags().StringP("repo_pattern", "r", "", "github repo pattern")
	bundleAppendGithubDependencyCommand.MarkFlagRequired("repo_pattern")

	bundleAppendMarketplaceDependencyCommand.Flags().StringP("marketplace_pattern", "m", "", "marketplace pattern")
	bundleAppendMarketplaceDependencyCommand.MarkFlagRequired("marketplace_pattern")

	bundleAppendPackageDependencyCommand.Flags().StringP("package_path", "p", "", "path to the package")
	bundleAppendPackageDependencyCommand.MarkFlagRequired("package_path")
}

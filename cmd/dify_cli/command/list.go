package command

import (
	"fmt"
	"os"

	"github.com/langgenius/dify-plugin-daemon/cmd/dify_cli/config"
	"github.com/spf13/cobra"
)

var ListCmd = &cobra.Command{
	Use:   "list",
	Short: "List all available tools",
	Long:  `List all available tools from the configured providers.`,
	Run:   runList,
}

func runList(cmd *cobra.Command, args []string) {
	cfg, err := config.Load()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: config not found, run 'dify init' first\n")
		os.Exit(1)
	}

	for _, provider := range cfg.Providers {
		fmt.Printf("[%s]\n", provider.Identity.Name)
		for _, tool := range provider.Tools {
			fmt.Printf("  %s - %s\n", tool.Identity.Name, tool.Description.Human.EnUS)
		}
	}
}

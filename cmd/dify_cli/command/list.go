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

	fmt.Printf("Available tools (provider: %s):\n", cfg.Env.Provider)
	for _, tool := range cfg.Tools {
		fmt.Printf("  %s - %s\n", tool.Identity.Name, tool.Description.Human.EnUS)
	}
}

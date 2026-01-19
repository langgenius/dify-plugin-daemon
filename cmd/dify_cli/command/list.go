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
	Long:  `List all available tools and tool references from the configured providers.`,
	Run:   runList,
}

func runList(cmd *cobra.Command, args []string) {
	cfg, err := config.Load()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: config not found, run 'dify init' first\n")
		os.Exit(1)
	}

	if len(cfg.Tools) > 0 {
		fmt.Println("Available tools:")
		for _, tool := range cfg.Tools {
			fmt.Printf("  %s (%s) - %s\n", tool.Identity.Name, tool.Identity.Provider, tool.Description.LLM)
		}
	}

	if len(cfg.ToolReferences) > 0 {
		if len(cfg.Tools) > 0 {
			fmt.Println()
		}
		fmt.Println("Tool references:")
		for _, ref := range cfg.ToolReferences {
			name := config.GetReferenceSymlinkName(&ref)
			fmt.Printf("  %s -> %s (%s)\n", name, ref.ToolName, ref.ToolProvider)
		}
	}

	if len(cfg.Tools) == 0 && len(cfg.ToolReferences) == 0 {
		fmt.Println("No tools or tool references defined in config")
	}
}

package command

import (
	"fmt"
	"os"

	"github.com/langgenius/dify-plugin-daemon/cmd/dify_cli/config"
	"github.com/langgenius/dify-plugin-daemon/cmd/dify_cli/tool"
	"github.com/langgenius/dify-plugin-daemon/cmd/dify_cli/types"
	"github.com/spf13/cobra"
)

var ListCmd = &cobra.Command{
	Use:   "list",
	Short: "List all available tools",
	Long:  `List all available tool references from the config.`,
	Run:   runList,
}

func runList(cmd *cobra.Command, args []string) {
	cfg, err := config.Load()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: config not found\n")
		os.Exit(1)
	}

	if len(cfg.ToolReferences) == 0 {
		fmt.Println("No tool references defined in config")
		return
	}

	missingRefs := findMissingTools(cfg)
	if len(missingRefs) > 0 {
		fmt.Fprintf(os.Stderr, "Fetching %d tool(s) from server...\n", len(missingRefs))
		tools, err := tool.FetchToolsBatch(cfg, missingRefs)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Warning: failed to fetch tools: %v\n", err)
		} else {
			cfg.Tools = append(cfg.Tools, tools...)
			if err := config.Save(cfg); err != nil {
				fmt.Fprintf(os.Stderr, "Warning: failed to save config: %v\n", err)
			}
		}
	}

	fmt.Println("Available tool references:")
	for _, ref := range cfg.ToolReferences {
		symName := config.GetReferenceSymlinkName(&ref)
		toolDecl := config.FindToolByReference(cfg, &ref)

		if toolDecl == nil {
			fmt.Printf("  %s - (not cached)\n", symName)
			continue
		}

		status := "[enabled]"
		if toolDecl.Enabled != nil && !*toolDecl.Enabled {
			status = "[disabled]"
		}

		desc := toolDecl.Description.LLM
		if desc == "" {
			desc = toolDecl.Description.Human.EnUS
		}
		if len(desc) > 60 {
			desc = desc[:57] + "..."
		}

		fmt.Printf("  %s %s - %s\n", status, symName, desc)
	}
}

func findMissingTools(cfg *types.DifyConfig) []types.ToolReference {
	var missing []types.ToolReference
	for _, ref := range cfg.ToolReferences {
		if config.FindToolByReference(cfg, &ref) == nil {
			missing = append(missing, ref)
		}
	}
	return missing
}

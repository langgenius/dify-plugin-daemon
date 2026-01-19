package command

import (
	"fmt"
	"os"

	"github.com/langgenius/dify-plugin-daemon/cmd/dify_cli/config"
	"github.com/langgenius/dify-plugin-daemon/cmd/dify_cli/tool"
	"github.com/langgenius/dify-plugin-daemon/cmd/dify_cli/types"
	"github.com/spf13/cobra"
)

var HelpCmd = &cobra.Command{
	Use:   "help [tool_name]",
	Short: "Show help for tools",
	Long:  `Show help information for available tools. Use without arguments to list all tools, or specify a tool name for detailed information.`,
	Run:   runHelp,
}

func runHelp(cmd *cobra.Command, args []string) {
	cfg, err := config.Load()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: config not found, run 'dify init' first\n")
		os.Exit(1)
	}

	if len(args) == 0 {
		printAllTools(cfg)
		return
	}

	invoker, err := tool.NewInvoker(cfg, args[0])
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	invoker.ShowHelp()
}

func printAllTools(cfg *types.DifyConfig) {
	fmt.Println("Available tools:")
	fmt.Println()

	for _, t := range cfg.Tools {
		desc := t.Description.LLM
		if desc == "" {
			desc = t.Description.Human.EnUS
		}
		if len(desc) > 60 {
			desc = desc[:57] + "..."
		}
		fmt.Printf("  %-30s %s\n", t.Identity.Name, desc)
	}

	fmt.Println()
	fmt.Println("Use 'dify help <tool>' for more information about a tool.")
}

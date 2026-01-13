package command

import (
	"fmt"
	"os"
	"strings"

	"github.com/langgenius/dify-plugin-daemon/cmd/dify_cli/config"
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

	toolName := args[0]
	tool := findTool(cfg, toolName)
	if tool == nil {
		fmt.Fprintf(os.Stderr, "Error: tool '%s' not found\n", toolName)
		fmt.Fprintf(os.Stderr, "Run 'dify help' to see available tools\n")
		os.Exit(1)
	}

	PrintToolHelp(tool)
}

func printAllTools(cfg *types.DifyConfig) {
	fmt.Println("Available tools:")
	fmt.Println()

	for _, tool := range cfg.Tools {
		desc := tool.Description.LLM
		if desc == "" {
			desc = tool.Description.Human.EnUS
		}
		if len(desc) > 60 {
			desc = desc[:57] + "..."
		}
		fmt.Printf("  %-30s %s\n", tool.Identity.Name, desc)
	}

	fmt.Println()
	fmt.Println("Use 'dify help <tool>' for more information about a tool.")
}

func findTool(cfg *types.DifyConfig, name string) *types.DifyToolDeclaration {
	for i := range cfg.Tools {
		if cfg.Tools[i].Identity.Name == name {
			return &cfg.Tools[i]
		}
	}
	return nil
}

func PrintToolHelp(tool *types.DifyToolDeclaration) {
	fmt.Printf("Tool: %s\n", tool.Identity.Name)
	fmt.Println()

	desc := tool.Description.LLM
	if desc == "" {
		desc = tool.Description.Human.EnUS
	}
	fmt.Println("Description:")
	fmt.Printf("  %s\n", desc)
	fmt.Println()

	if len(tool.Parameters) == 0 {
		fmt.Println("Parameters: None")
		return
	}

	fmt.Println("Parameters:")
	for _, param := range tool.Parameters {
		requiredStr := ""
		if param.Required {
			requiredStr = " (required)"
		}

		fmt.Printf("\n  --%s%s\n", param.Name, requiredStr)
		fmt.Printf("      Type: %s\n", param.Type)

		desc := param.LLMDescription
		if desc == "" {
			desc = param.HumanDescription.EnUS
		}
		if desc != "" {
			wrapped := wrapText(desc, 60)
			for i, line := range wrapped {
				if i == 0 {
					fmt.Printf("      Description: %s\n", line)
				} else {
					fmt.Printf("                   %s\n", line)
				}
			}
		}

		if param.Default != nil {
			fmt.Printf("      Default: %v\n", param.Default)
		}

		switch param.Type {
		case "select":
			if len(param.Options) > 0 {
				fmt.Printf("      Options:\n")
				for _, opt := range param.Options {
					label := opt.Label.EnUS
					if label == "" {
						label = opt.Value
					}
					if label == opt.Value {
						fmt.Printf("        - %s\n", opt.Value)
					} else {
						fmt.Printf("        - %s (%s)\n", opt.Value, label)
					}
				}
			}
		case "file":
			fmt.Printf("      Accepts: single file path (relative or absolute)\n")
		case "files":
			fmt.Printf("      Accepts: multiple file paths (relative or absolute, comma-separated)\n")
		case "boolean":
			fmt.Printf("      Values: true, false\n")
		case "number":
			fmt.Printf("      Values: numeric value\n")
		}
	}
}

func wrapText(text string, width int) []string {
	if len(text) <= width {
		return []string{text}
	}

	var lines []string
	words := strings.Fields(text)
	currentLine := ""

	for _, word := range words {
		if currentLine == "" {
			currentLine = word
		} else if len(currentLine)+1+len(word) <= width {
			currentLine += " " + word
		} else {
			lines = append(lines, currentLine)
			currentLine = word
		}
	}

	if currentLine != "" {
		lines = append(lines, currentLine)
	}

	return lines
}

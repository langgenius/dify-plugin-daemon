package tool

import (
	"fmt"
	"strings"

	"github.com/langgenius/dify-plugin-daemon/cmd/dify_cli/types"
)

func PrintHelp(tool *types.DifyToolDeclaration, fixedParams map[string]any) {
	fmt.Printf("Tool: %s\n", tool.Identity.Name)

	if tool.Enabled != nil && !*tool.Enabled {
		fmt.Println("Status: disabled by user")
		fmt.Println()
		fmt.Println("This tool has been disabled by the user. You are not allowed to use it.")
		return
	}

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
		fixedValue, isFixed := fixedParams[param.Name]

		requiredStr := ""
		if isFixed {
			requiredStr = " (fixed)"
		} else if param.Required {
			requiredStr = " (required)"
		}

		fmt.Printf("\n  --%s%s\n", param.Name, requiredStr)
		fmt.Printf("      Type: %s\n", param.Type)

		if isFixed {
			fmt.Printf("      Fixed: %v\n", fixedValue)
		}

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

		if !isFixed && param.Default != nil {
			fmt.Printf("      Default: %v\n", param.Default)
		}

		printParamTypeInfo(param)
	}
}

func printParamTypeInfo(param types.ToolParameter) {
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

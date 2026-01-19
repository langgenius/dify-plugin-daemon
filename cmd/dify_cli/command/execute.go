package command

import (
	"fmt"
	"os"

	"github.com/langgenius/dify-plugin-daemon/cmd/dify_cli/config"
	"github.com/langgenius/dify-plugin-daemon/cmd/dify_cli/tool"
	"github.com/spf13/cobra"
)

var ExecuteCmd = &cobra.Command{
	Use:                "execute <tool_name> [--param value ...]",
	Short:              "Execute a tool directly",
	Long:               `Execute a tool directly by name. Useful for debugging.`,
	Example:            `  dify execute google_search --query "hello world"`,
	Args:               cobra.MinimumNArgs(1),
	DisableFlagParsing: true,
	Run:                runExecute,
}

func runExecute(cmd *cobra.Command, args []string) {
	Invoke(args[0], args[1:])
}

func Invoke(name string, args []string) {
	cfg, err := config.Load()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: config not found, run 'dify init' first: %v\n", err)
		os.Exit(1)
	}

	invoker, err := tool.NewInvoker(cfg, name)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	if len(args) == 0 || args[0] == "--help" || args[0] == "-h" {
		invoker.ShowHelp()
		return
	}

	params, err := invoker.PrepareParams(args)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	err = tool.CallAPI(cfg, invoker.GetTool(), params, invoker.GetCredentialID())
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: API call failed: %v\n", err)
		os.Exit(1)
	}
}

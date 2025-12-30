package main

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/langgenius/dify-plugin-daemon/cmd/dify_cli/command"
	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "dify",
	Short: "BusyBox-style CLI for Dify tools",
	Long: `dify_cli is a BusyBox-style CLI for invoking Dify platform tools.

After initialization, you can use symlinked commands directly:
  google_search --query "hello world"`,
}

func main() {
	progName := filepath.Base(os.Args[0])

	// symlink invocation: google_search --query "hello"
	if progName != "dify" && progName != "dify_cli" {
		command.InvokeTool(progName, os.Args[1:])
		return
	}

	// normal invocation: dify init / dify execute
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func init() {
	rootCmd.AddCommand(command.InitCmd)
	rootCmd.AddCommand(command.ExecuteCmd)
	rootCmd.AddCommand(command.ListCmd)
}

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

	// Check if invoked via symlink to self (BusyBox style)
	if progName != "dify" && progName != "dify_cli" && isSymlinkToSelf() {
		command.InvokeTool(progName, os.Args[1:])
		return
	}

	// normal invocation: dify init / dify execute / dify list
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func isSymlinkToSelf() bool {
	invokedPath, err := filepath.EvalSymlinks(os.Args[0])
	if err != nil {
		return false
	}

	selfPath, err := os.Executable()
	if err != nil {
		return false
	}
	selfReal, err := filepath.EvalSymlinks(selfPath)
	if err != nil {
		return false
	}

	return invokedPath == selfReal
}

func init() {
	rootCmd.AddCommand(command.InitCmd)
	rootCmd.AddCommand(command.ExecuteCmd)
	rootCmd.AddCommand(command.ListCmd)
}

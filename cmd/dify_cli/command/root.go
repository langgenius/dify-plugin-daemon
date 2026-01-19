package command

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/spf13/cobra"
)

var RootCmd = &cobra.Command{
	Use:   "dify",
	Short: "BusyBox-style CLI for Dify",
	Long: `dify is a BusyBox-style CLI for invoking Dify platform tools.

After initialization, you can use symlinked commands directly:
  google_search --query "hello world"`,
}

func init() {
	RootCmd.SetHelpCommand(&cobra.Command{Hidden: true})
	RootCmd.CompletionOptions.DisableDefaultCmd = true
	RootCmd.AddCommand(InitCmd)
	RootCmd.AddCommand(ExecuteCmd)
	RootCmd.AddCommand(ListCmd)
	RootCmd.AddCommand(HelpCmd)
	RootCmd.AddCommand(PullCmd)
	RootCmd.AddCommand(EnvCmd)
}

func Execute() {
	progName := filepath.Base(os.Args[0])

	// Check if invoked via symlink to self (BusyBox style)
	if progName != "dify" && isSymlinkToSelf() {
		Invoke(progName, os.Args[1:])
		return
	}

	// normal invocation: dify init
	if err := RootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func isSymlinkToSelf() bool {
	invokedPath := os.Args[0]

	// try to find the full path of the invoked command
	if !filepath.IsAbs(invokedPath) {
		if p, err := exec.LookPath(invokedPath); err == nil {
			invokedPath = p
		} else {
			// LookPath failed, try to convert to absolute path
			if abs, err := filepath.Abs(invokedPath); err == nil {
				invokedPath = abs
			}
		}
	} else {
		// normalize absolute path
		if abs, err := filepath.Abs(invokedPath); err == nil {
			invokedPath = abs
		}
	}

	// check if the file is a symlink using Lstat (doesn't follow symlinks)
	info, err := os.Lstat(invokedPath)
	if err != nil {
		return false
	}

	if info.Mode()&os.ModeSymlink == 0 {
		return false
	}

	// resolve symlink target and get its file info
	linkTarget, err := filepath.EvalSymlinks(invokedPath)
	if err != nil {
		return false
	}
	targetInfo, err := os.Stat(linkTarget)
	if err != nil {
		return false
	}

	// get current executable's file info
	selfPath, err := os.Executable()
	if err != nil {
		return false
	}
	selfInfo, err := os.Stat(selfPath)
	if err != nil {
		return false
	}

	// use os.SameFile to compare by inode/file ID
	return os.SameFile(targetInfo, selfInfo)
}

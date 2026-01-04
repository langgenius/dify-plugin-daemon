package command

import (
	"fmt"
	"os"

	"github.com/langgenius/dify-plugin-daemon/cmd/dify_cli/config"
	"github.com/spf13/cobra"
)

var InitCmd = &cobra.Command{
	Use:   "init",
	Short: "Initialize tool symlinks from config",
	Long:  `Create symlinks for all tools defined in .dify_cli.json`,
	Run:   runInit,
}

func runInit(cmd *cobra.Command, args []string) {
	cfg, err := config.Load()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: failed to load %s: %v\n", config.GetConfigPath(), err)
		os.Exit(1)
	}

	if len(cfg.Tools) == 0 {
		fmt.Fprintf(os.Stderr, "No tools defined in config\n")
		return
	}

	selfPath, err := config.GetSelfPath()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: failed to get executable path: %v\n", err)
		os.Exit(1)
	}

	created := 0
	skipped := 0
	for _, tool := range cfg.Tools {
		name := tool.Identity.Name
		if _, err := os.Lstat(name); err == nil {
			fmt.Fprintf(os.Stdout, "  [SKIP] %s (already exists)\n", name)
			skipped++
			continue
		}
		if err := os.Symlink(selfPath, name); err != nil {
			fmt.Fprintf(os.Stderr, "  [FAIL] %s: %v\n", name, err)
			continue
		}
		fmt.Fprintf(os.Stdout, "  [OK] %s\n", name)
		created++
	}

	fmt.Fprintf(os.Stdout, "\nCreated %d symlinks, skipped %d\n", created, skipped)
	fmt.Fprintf(os.Stdout, "\nRun this to add tools to PATH:\n  eval \"$(dify env)\"\n")
}

package command

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/langgenius/dify-plugin-daemon/cmd/dify_cli/config"
	"github.com/langgenius/dify-plugin-daemon/cmd/dify_cli/types"
	"github.com/spf13/cobra"
)

var (
	envFile    string
	schemaFile string
)

var InitCmd = &cobra.Command{
	Use:   "init",
	Short: "Initialize with config files",
	Long: `Initialize dify_cli with environment and tool schema files.

This command reads the env file and schema file, saves the configuration,
and creates symlinks for each tool in ~/.dify/bin/.`,
	Example: `  dify init --env /path/to/dify.env --schemas /path/to/tools_schema.json`,
	Run:     runInit,
}

func init() {
	InitCmd.Flags().StringVar(&envFile, "env", "", "Path to environment file (required)")
	InitCmd.Flags().StringVar(&schemaFile, "schemas", "", "Path to tool schema JSON file (required)")
	InitCmd.MarkFlagRequired("env")
	InitCmd.MarkFlagRequired("schemas")
}

func runInit(cmd *cobra.Command, args []string) {
	env, err := config.LoadEnvFile(envFile)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: failed to load env file: %v\n", err)
		os.Exit(1)
	}

	if env.InnerAPIURL == "" || env.InnerAPIKey == "" {
		fmt.Fprintf(os.Stderr, "Error: env file must contain INNER_API_URL and INNER_API_KEY\n")
		os.Exit(1)
	}

	schemas, err := config.LoadSchemaFile(schemaFile)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: failed to load schema file: %v\n", err)
		os.Exit(1)
	}

	cfg := &types.DifyConfig{
		Env:   env,
		Tools: schemas.Tools,
	}

	if err := config.Save(cfg); err != nil {
		fmt.Fprintf(os.Stderr, "Error: failed to save config: %v\n", err)
		os.Exit(1)
	}

	selfPath, err := config.GetSelfPath()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: failed to get executable path: %v\n", err)
		os.Exit(1)
	}

	binDir := config.GetBinDir()
	if err := os.MkdirAll(binDir, 0755); err != nil {
		fmt.Fprintf(os.Stderr, "Error: failed to create bin directory: %v\n", err)
		os.Exit(1)
	}

	created := 0
	for _, tool := range schemas.Tools {
		linkPath := filepath.Join(binDir, tool.Identity.Name)

		os.Remove(linkPath)

		if err := os.Symlink(selfPath, linkPath); err != nil {
			fmt.Printf("  [SKIP] %s: %v\n", tool.Identity.Name, err)
			continue
		}
		fmt.Printf("  [OK] %s\n", tool.Identity.Name)
		created++
	}

	fmt.Println()
	fmt.Printf("Initialized %d tools in %s\n", created, binDir)
	fmt.Println()
	fmt.Println("Add to your PATH:")
	fmt.Printf("  export PATH=\"%s:$PATH\"\n", binDir)
}

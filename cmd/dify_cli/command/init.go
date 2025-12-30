package command

import (
	"fmt"
	"os"

	"github.com/langgenius/dify-plugin-daemon/cmd/dify_cli/config"
	"github.com/langgenius/dify-plugin-daemon/cmd/dify_cli/types"
	"github.com/langgenius/dify-plugin-daemon/pkg/validators"
	"github.com/spf13/cobra"
)

var (
	envFile    string
	schemaFile string
)

var InitCmd = &cobra.Command{
	Use:     "init",
	Short:   "Initialize with config files",
	Example: `  dify init --env dify.env --schemas tools_schema.json`,
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
	if err := validators.GlobalEntitiesValidator.Struct(env); err != nil {
		fmt.Fprintf(os.Stderr, "Error: invalid env config: %v\n", err)
		os.Exit(1)
	}

	schemas, err := config.LoadSchemaFile(schemaFile)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: failed to load schema file: %v\n", err)
		os.Exit(1)
	}

	if err := config.Save(&types.DifyConfig{Env: env, Tools: schemas.Tools}); err != nil {
		fmt.Fprintf(os.Stderr, "Error: failed to save config: %v\n", err)
		os.Exit(1)
	}

	selfPath, err := config.GetSelfPath()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: failed to get executable path: %v\n", err)
		os.Exit(1)
	}

	created := 0
	for _, tool := range schemas.Tools {
		name := tool.Identity.Name
		os.Remove(name)
		if err := os.Symlink(selfPath, name); err != nil {
			fmt.Printf("  [SKIP] %s: %v\n", name, err)
			continue
		}
		fmt.Printf("  [OK] %s\n", name)
		created++
	}

	fmt.Printf("\nInitialized %d tools in current directory\n", created)
}

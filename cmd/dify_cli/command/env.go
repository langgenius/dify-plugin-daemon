package command

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var EnvCmd = &cobra.Command{
	Use:   "env",
	Short: "Output shell commands to add tools to PATH",
	Long:  `Output export command for eval. Usage: eval "$(dify env)"`,
	Run:   runEnv,
}

func runEnv(cmd *cobra.Command, args []string) {
	dir, err := os.Getwd()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
	fmt.Printf("export PATH=\"%s:$PATH\"\n", dir)
}

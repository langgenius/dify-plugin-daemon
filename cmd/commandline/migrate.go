package main

import (
	"fmt"
	"os"

	"github.com/joho/godotenv"
	"github.com/kelseyhightower/envconfig"
	"github.com/langgenius/dify-plugin-daemon/internal/db"
	"github.com/langgenius/dify-plugin-daemon/internal/types/app"
	"github.com/spf13/cobra"
)

var migrateCommand = &cobra.Command{
	Use:   "migrate",
	Short: "Run database migrations",
	Long:  "Run database auto migration for the Dify Plugin Daemon",
	Run: func(cmd *cobra.Command, args []string) {
		_ = godotenv.Load()

		var config app.Config
		if err := envconfig.Process("", &config); err != nil {
			fmt.Fprintf(os.Stderr, "error processing environment variables: %v\n", err)
			os.Exit(1)
		}

		config.SetDefault()

		if err := config.Validate(); err != nil {
			fmt.Fprintf(os.Stderr, "invalid configuration: %v\n", err)
			os.Exit(1)
		}

		db.Init(&config)
		defer db.Close()

		if err := db.AutoMigrate(); err != nil {
			fmt.Fprintf(os.Stderr, "failed to run migrations: %v\n", err)
			os.Exit(1)
		}

		fmt.Println("database migration completed successfully")
	},
}

package main

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var (
	cfgFile string

	rootCommand = &cobra.Command{
		Use:   "dify",
		Short: "Dify",
		Long:  "Dify is a cli tool to help you develop your Dify projects.",
	}

	pluginCommand = &cobra.Command{
		Use:   "plugin",
		Short: "Plugin",
		Long:  "Plugin related commands",
	}

	bundleCommand = &cobra.Command{
		Use:   "bundle",
		Short: "Bundle",
		Long:  "Bundle related commands",
	}
)

func init() {
	cobra.OnInitialize(initConfig)

	rootCommand.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is $HOME/.dify.yaml)")
	rootCommand.AddCommand(pluginCommand)
	rootCommand.AddCommand(bundleCommand)
}

func initConfig() {
	if cfgFile != "" {
		// Use config file from the flag.
		viper.SetConfigFile(cfgFile)
	} else {
		// Find home directory.
		home, err := os.UserHomeDir()
		cobra.CheckErr(err)

		// Search config in home directory with name ".dify" (without extension).
		viper.AddConfigPath(home)
		viper.SetConfigType("yaml")
		viper.SetConfigName(".dify")
	}

	viper.AutomaticEnv()

	if err := viper.ReadInConfig(); err == nil {
		fmt.Println("Using config file:", viper.ConfigFileUsed())
	}
}

func main() {
	rootCommand.Execute()
}

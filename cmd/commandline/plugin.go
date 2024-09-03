package main

import (
	"fmt"
	"os"

	init_pkg "github.com/langgenius/dify-plugin-daemon/cmd/commandline/init"
	"github.com/langgenius/dify-plugin-daemon/internal/core/plugin_packager/checksum"
	"github.com/langgenius/dify-plugin-daemon/internal/core/plugin_packager/decoder"
	"github.com/langgenius/dify-plugin-daemon/internal/core/plugin_packager/packager"
	"github.com/langgenius/dify-plugin-daemon/internal/utils/log"
	"github.com/spf13/cobra"
)

var (
	pluginInitCommand = &cobra.Command{
		Use:   "init",
		Short: "Init",
		Long:  "Init",
		Run: func(c *cobra.Command, args []string) {
			init_pkg.InitPlugin()
		},
	}

	pluginPackageCommand = &cobra.Command{
		Use:   "package plugin_path [-o output_path]",
		Short: "Package",
		Long:  "Package plugins",
		Run: func(cmd *cobra.Command, args []string) {
			if len(args) < 1 {
				fmt.Println("Error: plugin_path is required")
				return
			}
			output_path := "./plugin.difypkg"
			if cmd.Flag("output_path") != nil {
				output_path = cmd.Flag("output_path").Value.String()
			}
			decoder, err := decoder.NewFSPluginDecoder(args[0])
			if err != nil {
				log.Error("failed to create plugin decoder , plugin path: %s, error: %v", args[0], err)
				return
			}

			packager := packager.NewPackager(decoder)
			zip_file, err := packager.Pack()

			if err != nil {
				log.Error("failed to package plugin %v", err)
				return
			}

			err = os.WriteFile(output_path, zip_file, 0644)
			if err != nil {
				log.Error("failed to write package file %v", err)
				return
			}

			log.Info("plugin packaged successfully, output path: %s", output_path)
		},
	}

	pluginChecksumCommand = &cobra.Command{
		Use:   "checksum plugin_path",
		Short: "Checksum",
		Long:  "Calculate the checksum of the plugin, you need specify the plugin path or .difypkg file path",
		Run: func(cmd *cobra.Command, args []string) {
			if len(args) < 1 {
				fmt.Println("Error: plugin_path is required")
				return
			}

			plugin_path := args[0]
			var plugin_decoder decoder.PluginDecoder
			if stat, err := os.Stat(plugin_path); err == nil {
				if stat.IsDir() {
					plugin_decoder, err = decoder.NewFSPluginDecoder(plugin_path)
					if err != nil {
						log.Error("failed to create plugin decoder, plugin path: %s, error: %v", plugin_path, err)
						return
					}
				} else {
					bytes, err := os.ReadFile(plugin_path)
					if err != nil {
						log.Error("failed to read plugin file, plugin path: %s, error: %v", plugin_path, err)
						return
					}

					plugin_decoder, err = decoder.NewZipPluginDecoder(bytes)
					if err != nil {
						log.Error("failed to create plugin decoder, plugin path: %s, error: %v", plugin_path, err)
						return
					}
				}
			} else {
				log.Error("failed to get plugin file info, plugin path: %s, error: %v", plugin_path, err)
				return
			}

			checksum, err := checksum.CalculateChecksum(plugin_decoder)
			if err != nil {
				log.Error("failed to calculate checksum, plugin path: %s, error: %v", plugin_path, err)
				return
			}

			log.Info("plugin checksum: %s", checksum)
		},
	}

	pluginPermissionCommand = &cobra.Command{
		Use:   "permission",
		Short: "Permission",
		Long: `Permission, available values: 
tools					- allow plugin to call tools
models					- allow plugin to call models
models.llm				- allow plugin to call llm
models.text_embedding			- allow plugin to call text_embedding model
models.rerank				- allow plugin to call rerank model
models.tts				- allow plugin to call tts
models.speech2text			- allow plugin to call speech2text
models.moderation			- allow plugin to call moderation
apps					- allow plugin to call apps
storage					- allow plugin to use storage
endpoint				- allow plugin to register endpoint`,
	}

	pluginPermissionAddCommand = &cobra.Command{
		Use:   "add permission",
		Short: "",
		Long:  "Add permission to plugin, you can find the available permission by running `dify plugin permission`",
	}

	pluginPermissionDropCommand = &cobra.Command{
		Use:   "drop permission",
		Short: "",
		Long:  "Drop permission from plugin, you can find the available permission by running `dify plugin permission`",
	}
)

func init() {
	pluginCommand.AddCommand(pluginInitCommand)
	pluginCommand.AddCommand(pluginPackageCommand)
	pluginCommand.AddCommand(pluginChecksumCommand)
	pluginCommand.AddCommand(pluginPermissionCommand)
	pluginPermissionCommand.AddCommand(pluginPermissionAddCommand)
	pluginPermissionCommand.AddCommand(pluginPermissionDropCommand)
}

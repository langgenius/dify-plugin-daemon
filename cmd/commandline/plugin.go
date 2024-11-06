package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	init_pkg "github.com/langgenius/dify-plugin-daemon/cmd/commandline/init"
	"github.com/langgenius/dify-plugin-daemon/internal/core/plugin_daemon/access_types"
	"github.com/langgenius/dify-plugin-daemon/internal/core/plugin_manager"
	"github.com/langgenius/dify-plugin-daemon/internal/core/plugin_packager/decoder"
	"github.com/langgenius/dify-plugin-daemon/internal/core/plugin_packager/packager"
	"github.com/langgenius/dify-plugin-daemon/internal/types/app"
	"github.com/langgenius/dify-plugin-daemon/internal/utils/log"
	"github.com/langgenius/dify-plugin-daemon/internal/utils/parser"
	"github.com/langgenius/dify-plugin-daemon/internal/utils/routine"
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
			input_path := args[0]
			// using filename of input_path as output_path if not specified
			output_path := ""

			if cmd.Flag("output_path") != nil {
				output_path = cmd.Flag("output_path").Value.String()
			} else {
				output_path = filepath.Base(input_path) + ".difypkg"
			}

			decoder, err := decoder.NewFSPluginDecoder(input_path)
			if err != nil {
				log.Error("failed to create plugin decoder , plugin path: %s, error: %v", input_path, err)
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

			checksum, err := plugin_decoder.Checksum()
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

	pluginTestCommand = &cobra.Command{
		Use:   "test [-i inputs] [-t timeout] package_path invoke_type invoke_action",
		Short: "",
		Long: "Test runs the given plugin package locally, and you can specify the inputs using json format, if not specified, will use default inputs\n" +
			"type: invoke type, available values: \n" +
			"[\n" +
			"	tool, model, endpoint\n" +
			"]\n" +
			"action: invoke action, available values: \n" +
			"[\n" +
			"	invoke_tool, validate_tool_credentials, \n" +
			"	invoke_endpoint\n" +
			"	invoke_llm, invoke_text_embedding, invoke_rerank, invoke_tts, invoke_speech2text, invoke_moderation, \n" +
			"	validate_provider_credentials, validate_model_credentials, get_tts_model_voices, \n" +
			"	get_text_embedding_num_tokens, get_ai_model_schemas, get_llm_num_tokens\n" +
			"]\n",
		Run: func(cmd *cobra.Command, args []string) {
			if len(args) < 3 {
				log.Error("invalid args, please specify package_path, invoke_type, invoke_action")
				return
			}
			// get package path
			package_path_str := args[0]
			// get invoke type
			invoke_type_str := args[1]
			// get invoke action
			invoke_action_str := args[2]
			// get inputs if specified
			inputs := map[string]any{}
			if cmd.Flag("inputs") != nil {
				inputs_str := cmd.Flag("inputs").Value.String()
				err := json.Unmarshal([]byte(inputs_str), &inputs)
				if err != nil {
					log.Error("failed to unmarshal inputs, inputs: %s, error: %v", inputs_str, err)
					return
				}
			}
			// parse flag
			timeout := ""
			if cmd.Flag("timeout") != nil {
				timeout = cmd.Flag("timeout").Value.String()
			}

			// get invoke_type and invoke_action
			invoke_type := access_types.PluginAccessType(invoke_type_str)
			if !invoke_type.IsValid() {
				log.Error("invalid invoke type: %s", invoke_type_str)
				return
			}
			invoke_action := access_types.PluginAccessAction(invoke_action_str)
			if !invoke_action.IsValid() {
				log.Error("invalid invoke action: %s", invoke_action_str)
				return
			}

			// init routine pool
			routine.InitPool(1024)

			// clean working directory when test finished
			defer os.RemoveAll("./working")

			// init testing config
			config := &app.Config{
				PluginWorkingPath:    "./working/cwd",
				PluginStoragePath:    "./working/storage",
				PluginMediaCachePath: "./working/media_cache",
				ProcessCachingPath:   "./working/subprocesses",
				Platform:             app.PLATFORM_LOCAL,
			}
			config.SetDefault()

			// init plugin manager
			plugin_manager := plugin_manager.InitGlobalManager(config)

			response, err := plugin_manager.TestPlugin(package_path_str, inputs, invoke_type, invoke_action, timeout)
			if err != nil {
				log.Error("failed to test plugin, package_path: %s, error: %v", package_path_str, err)
				return
			}

			for response.Next() {
				item, err := response.Read()
				if err != nil {
					log.Error("failed to read response item, error: %v", err)
					return
				}
				log.Info("%v", parser.MarshalJson(item))
			}
		},
	}
)

func init() {
	pluginCommand.AddCommand(pluginInitCommand)
	pluginCommand.AddCommand(pluginPackageCommand)
	pluginCommand.AddCommand(pluginChecksumCommand)
	pluginCommand.AddCommand(pluginPermissionCommand)
	pluginCommand.AddCommand(pluginTestCommand)
	pluginTestCommand.Flags().StringP("inputs", "i", "", "inputs")
	pluginTestCommand.Flags().StringP("timeout", "t", "", "timeout")
	pluginPermissionCommand.AddCommand(pluginPermissionAddCommand)
	pluginPermissionCommand.AddCommand(pluginPermissionDropCommand)
}
